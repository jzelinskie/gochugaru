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

func (c *Client) CheckOne(ctx context.Context, r Relationship) (bool, error) {
	var b CheckBuilder
	b.AddRelationship(r)
	results, err := c.Check(ctx, &b)
	if err != nil {
		return false, err
	}
	return results[0], nil
}

func (c *Client) CheckAny(ctx context.Context, b *CheckBuilder) (bool, error) {
	results, err := c.Check(ctx, b)
	if err != nil {
		return false, err
	}

	return slices.Contains(results, true), nil
}

func (c *Client) CheckAll(ctx context.Context, b *CheckBuilder) (bool, error) {
	results, err := c.Check(ctx, b)
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

func (c *Client) Check(ctx context.Context, b *CheckBuilder) ([]bool, error) {
	resp, err := c.client.BulkCheckPermission(ctx, &v1.BulkCheckPermissionRequest{
		Consistency: b.consistency,
		Items:       b.items,
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

func (c *Client) ForEachRelationship(ctx context.Context, f *Filter, fn RelationshipFunc) error {
	stream, err := c.client.ReadRelationships(ctx, &v1.ReadRelationshipsRequest{
		Consistency:        &v1.Consistency{}, // TODO(jzelinskie): decide API for consistency
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
