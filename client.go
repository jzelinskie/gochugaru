// Package gochugaru implements an ergonomic SpiceDB client that wraps the
// official AuthZed gRPC client.
package gochugaru

import (
	"context"
	"errors"
	"io"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "github.com/mostynb/go-grpc-compression/experimental/s2" // Register Snappy S2 compression
)

var defaultClientOpts = []grpc.DialOption{
	grpc.WithDefaultCallOptions(grpc.UseCompressor("s2")),
}

// NewPlaintextClient creates a client that does not enforce TLS.
//
// This should be used only for testing (usually against localhost).
func NewPlaintextClient(endpoint, presharedKey string) (*Client, error) {
	return NewClientWithOpts(endpoint, append(
		defaultClientOpts,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpcutil.WithInsecureBearerToken(presharedKey),
	)...)
}

// NewSystemTLSClient creates a client using TLS verified by the operating
// system's certificate chain.
//
// This should be sufficient for production usage in the typical environments.
func NewSystemTLSClient(endpoint, presharedKey string) (*Client, error) {
	withSystemCerts, err := grpcutil.WithSystemCerts(grpcutil.VerifyCA)
	if err != nil {
		return nil, err
	}

	return NewClientWithOpts(endpoint, append(
		defaultClientOpts,
		withSystemCerts,
		grpcutil.WithBearerToken("t_your_token_here_1234567deadbeef"),
	)...)
}

// NewClientWithOpts creates a client that allows for configuring gRPC options.
//
// This should only be used if the other methods don't suffice.
//
// I'd love to hear about what DialOptions you're using in the SpiceDB Discord
// (https://discord.gg/spicedb) or the issue tracker for this library.
func NewClientWithOpts(endpoint string, opts ...grpc.DialOption) (*Client, error) {
	client, err := authzed.NewClientWithExperimentalAPIs(endpoint, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{client: client}, nil
}

type Client struct {
	client *authzed.ClientWithExperimental
}

// Write atomically performs a transaction on relationships.
func (c *Client) Write(ctx context.Context, txn *Txn) (writtenAtRevision string, err error) {
	resp, err := c.client.WriteRelationships(ctx, &v1.WriteRelationshipsRequest{
		Updates:               txn.updates,
		OptionalPreconditions: txn.preconds,
	})
	if err != nil {
		return "", err
	}
	return resp.WrittenAt.Token, nil
}

// CheckOne performs a permissions check for a single relationship.
func (c *Client) CheckOne(ctx context.Context, cs *Consistency, r Relationshipper) (bool, error) {
	results, err := c.Check(ctx, cs, r)
	if err != nil {
		return false, err
	}
	return results[0], nil
}

// CheckAny returns true if any of the provided relationships have access.
func (c *Client) CheckAny(ctx context.Context, cs *Consistency, rs []Relationshipper) (bool, error) {
	results, err := c.Check(ctx, cs, rs...)
	if err != nil {
		return false, err
	}

	return slices.Contains(results, true), nil
}

// CheckAll returns true if all of the provided relationships have access.
func (c *Client) CheckAll(ctx context.Context, cs *Consistency, rs []Relationshipper) (bool, error) {
	results, err := c.Check(ctx, cs, rs...)
	if err != nil {
		return false, err
	}

	for _, result := range results {
		if !result {
			return false, nil
		}
	}
	return true, nil
}

// Check performs a batched permissions check for the provided relationships.
func (c *Client) Check(ctx context.Context, cs *Consistency, rs ...Relationshipper) ([]bool, error) {
	items := make([]*v1.BulkCheckPermissionRequestItem, 0, len(rs))
	for _, ir := range rs {
		r := ir.Relationship()
		items = append(items, &v1.BulkCheckPermissionRequestItem{
			Resource: &v1.ObjectReference{
				ObjectType: r.ResourceType,
				ObjectId:   r.ResourceID,
			},
			Permission: r.ResourceRelation,
			Subject: &v1.SubjectReference{
				Object: &v1.ObjectReference{
					ObjectType: r.SubjectType,
					ObjectId:   r.SubjectRelation,
				},
				OptionalRelation: r.SubjectRelation,
			},
			Context: r.caveat().GetContext(),
		})
	}

	resp, err := c.client.BulkCheckPermission(ctx, &v1.BulkCheckPermissionRequest{
		Consistency: cs.v1c,
		Items:       items,
	})
	if err != nil {
		return nil, err
	}

	var results []bool
	for _, pair := range resp.Pairs {
		switch resp := pair.Response.(type) {
		case *v1.BulkCheckPermissionPair_Item:
			results = append(
				results,
				resp.Item.Permissionship == v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION,
			)
		case *v1.BulkCheckPermissionPair_Error:
			return results, errors.New(resp.Error.Message)
		}
	}
	return results, nil
}

// ForEachRelationship calls the provided function for each relationship
// matching the provided filter.
func (c *Client) ForEachRelationship(ctx context.Context, cs *Consistency, f *Filter, fn RelationshipFunc) error {
	stream, err := c.client.ReadRelationships(ctx, &v1.ReadRelationshipsRequest{
		Consistency:        cs.v1c,
		RelationshipFilter: f.filter,
		// TODO(jzelinskie): handle pagination for folks
	})
	if err != nil {
		return err
	}

	for resp, err := stream.Recv(); err != io.EOF; resp, err = stream.Recv() {
		if err != nil {
			return err
		}

		if err := fn(fromV1(resp.Relationship)); err != nil {
			return err
		}
	}

	return nil
}

// DeleteAtomic removes all of the relationships matching the provided filter
// in a single transaction.
func (c *Client) DeleteAtomic(ctx context.Context, f *PreconditionedFilter) (deletedAtRevision string, err error) {
	resp, err := c.client.DeleteRelationships(ctx, &v1.DeleteRelationshipsRequest{
		RelationshipFilter:            f.filter,
		OptionalPreconditions:         f.preconds,
		OptionalLimit:                 0,
		OptionalAllowPartialDeletions: false,
	})
	if err != nil {
		return "", err
	} else if resp.DeletionProgress != v1.DeleteRelationshipsResponse_DELETION_PROGRESS_COMPLETE {
		return "", errors.New("delete disallowing partial deletion did not complete")
	}

	return resp.DeletedAt.Token, nil
}

// Delete removes all of the relationships matching the provided filter in
// batches.
func (c *Client) Delete(ctx context.Context, f *PreconditionedFilter) error {
	for {
		resp, err := c.client.DeleteRelationships(ctx, &v1.DeleteRelationshipsRequest{
			RelationshipFilter:            f.filter,
			OptionalPreconditions:         f.preconds,
			OptionalLimit:                 10_000,
			OptionalAllowPartialDeletions: false,
		})
		if err != nil {
			return err
		} else if resp.DeletionProgress == v1.DeleteRelationshipsResponse_DELETION_PROGRESS_COMPLETE {
			break
		}
	}
	return nil
}
