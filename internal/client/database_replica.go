package client

import (
	"context"
	"fmt"
	"time"
)

// DatabaseReplica represents a read replica of a database instance.
type DatabaseReplica struct {
	Name                 string  `json:"name"`
	NodeID               string  `json:"node_id"`
	ReplicaIndex         int     `json:"replica_index"`
	Endpoint             *string `json:"endpoint"`
	Status               string  `json:"status"`
	Ready                bool    `json:"ready"`
	ReplicationStatus    *string `json:"replication_status"`
	SecondsBehindMaster  *int    `json:"seconds_behind_master"`
	IsReplicationHealthy bool    `json:"is_replication_healthy"`
}

// DatabaseReplicaMaster represents the master node in a replica listing.
type DatabaseReplicaMaster struct {
	Name     string  `json:"name"`
	NodeID   string  `json:"node_id"`
	Endpoint *string `json:"endpoint"`
	Status   string  `json:"status"`
	Ready    bool    `json:"ready"`
}

// DatabaseReplicaList is the response from listing replicas for an instance.
type DatabaseReplicaList struct {
	Replicas []DatabaseReplica     `json:"replicas"`
	Master   DatabaseReplicaMaster `json:"master"`
	Billing  struct {
		HourlyCostCents  int `json:"hourly_cost_cents"`
		MonthlyCostCents int `json:"monthly_cost_cents"`
	} `json:"billing"`
}

// AddDatabaseReplicasRequest is the payload for adding replicas.
type AddDatabaseReplicasRequest struct {
	ReplicaCount int `json:"replica_count"`
}

type addDatabaseReplicasResponse struct {
	Message  string            `json:"message"`
	Replicas []DatabaseReplica `json:"replicas"`
}

// ListDatabaseReplicas returns the master + replicas for a database instance.
func (c *Client) ListDatabaseReplicas(ctx context.Context, instanceID string) (*DatabaseReplicaList, error) {
	var resp DatabaseReplicaList
	if err := c.doRequest(ctx, "GET", fmt.Sprintf("/database/%s/replicas", instanceID), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AddDatabaseReplicas adds one or more replicas to a database instance.
func (c *Client) AddDatabaseReplicas(ctx context.Context, instanceID string, req AddDatabaseReplicasRequest) ([]DatabaseReplica, error) {
	var resp addDatabaseReplicasResponse
	if err := c.doRequest(ctx, "POST", fmt.Sprintf("/database/%s/replicas", instanceID), req, &resp); err != nil {
		return nil, err
	}
	return resp.Replicas, nil
}

// DeleteDatabaseReplica removes the replica at the given index.
func (c *Client) DeleteDatabaseReplica(ctx context.Context, instanceID string, replicaIndex int) error {
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("/database/%s/replicas/%d", instanceID, replicaIndex), nil, nil)
}

// FindDatabaseReplica returns the replica at the given index (or NotFoundError).
func (c *Client) FindDatabaseReplica(ctx context.Context, instanceID string, replicaIndex int) (*DatabaseReplica, error) {
	list, err := c.ListDatabaseReplicas(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	for i := range list.Replicas {
		if list.Replicas[i].ReplicaIndex == replicaIndex {
			return &list.Replicas[i], nil
		}
	}
	return nil, &NotFoundError{Resource: "database replica", ID: fmt.Sprintf("%s:%d", instanceID, replicaIndex)}
}

// WaitForDatabaseReplicaReady waits for a replica to become ready.
func (c *Client) WaitForDatabaseReplicaReady(ctx context.Context, instanceID string, replicaIndex int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for database replica %s:%d to become ready", instanceID, replicaIndex)
			}
			replica, err := c.FindDatabaseReplica(ctx, instanceID, replicaIndex)
			if err != nil {
				if IsNotFound(err) {
					continue
				}
				return fmt.Errorf("error checking database replica status: %w", err)
			}
			if replica.Ready {
				return nil
			}
			if replica.Status == "error" || replica.Status == "failed" {
				return fmt.Errorf("database replica %s:%d entered error state", instanceID, replicaIndex)
			}
		}
	}
}
