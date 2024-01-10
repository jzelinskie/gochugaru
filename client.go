// Package gochugaru implements an ergonomic SpiceDB client that wraps the
// official AuthZed gRPC client.
package gochugaru

import (
	"context"
	"errors"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc"
)

func NewClient(endpoint string, opts ...grpc.DialOption) (*Client, error) {
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

func (c *Client) CheckOne(ctx context.Context, object, permission, subject string) (bool, error) {
	var b CheckBuilder
	b.AddRelationship(object, permission, subject)
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
