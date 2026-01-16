// Package node provides node management functionality for multi-server management.
package node

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"v/internal/database/repository"
	"v/internal/logger"
)

// Common errors
var (
	ErrNodeNotFound       = errors.New("node not found")
	ErrInvalidNode        = errors.New("invalid node data")
	ErrInvalidAddress     = errors.New("invalid node address")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenRevoked       = errors.New("token has been revoked")
	ErrNoHealthyNodes     = errors.New("no healthy nodes available")
	ErrNodeAtCapacity     = errors.New("node is at capacity")
	ErrDuplicateToken     = errors.New("duplicate token generated")
)

// TokenLength is the length of generated tokens in bytes (64 hex chars = 32 bytes)
const TokenLength = 32

// Node represents a node in the service layer.
type Node struct {
	ID           int64      `json:"id"`
	Name         string     `json:"name"`
	Address      string     `json:"address"`
	Port         int        `json:"port"`
	Token        string     `json:"token,omitempty"`
	Status       string     `json:"status"`
	Tags         []string   `json:"tags"`
	Region       string     `json:"region"`
	Weight       int        `json:"weight"`
	MaxUsers     int        `json:"max_users"`
	CurrentUsers int        `json:"current_users"`
	Latency      int        `json:"latency"`
	LastSeenAt   *time.Time `json:"last_seen_at"`
	SyncStatus   string     `json:"sync_status"`
	SyncedAt     *time.Time `json:"synced_at"`
	IPWhitelist  []string   `json:"ip_whitelist"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// NodeMetrics represents node metrics for updates.
type NodeMetrics struct {
	Latency      int `json:"latency"`
	CurrentUsers int `json:"current_users"`
}

// CreateNodeRequest represents a request to create a node.
type CreateNodeRequest struct {
	Name        string   `json:"name"`
	Address     string   `json:"address"`
	Port        int      `json:"port"`
	Tags        []string `json:"tags"`
	Region      string   `json:"region"`
	Weight      int      `json:"weight"`
	MaxUsers    int      `json:"max_users"`
	IPWhitelist []string `json:"ip_whitelist"`
}

// UpdateNodeRequest represents a request to update a node.
type UpdateNodeRequest struct {
	Name        *string   `json:"name"`
	Address     *string   `json:"address"`
	Port        *int      `json:"port"`
	Tags        *[]string `json:"tags"`
	Region      *string   `json:"region"`
	Weight      *int      `json:"weight"`
	MaxUsers    *int      `json:"max_users"`
	IPWhitelist *[]string `json:"ip_whitelist"`
}

// NodeFilter defines filter options for listing nodes.
type NodeFilter struct {
	Status  string
	Region  string
	Tags    []string
	GroupID *int64
	Limit   int
	Offset  int
}


// Service provides node management operations.
type Service struct {
	nodeRepo           repository.NodeRepository
	assignmentRepo     repository.UserNodeAssignmentRepository
	logger             logger.Logger
}

// NewService creates a new node service.
func NewService(
	nodeRepo repository.NodeRepository,
	assignmentRepo repository.UserNodeAssignmentRepository,
	log logger.Logger,
) *Service {
	return &Service{
		nodeRepo:       nodeRepo,
		assignmentRepo: assignmentRepo,
		logger:         log,
	}
}

// Create creates a new node.
func (s *Service) Create(ctx context.Context, req *CreateNodeRequest) (*Node, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidNode)
	}
	if req.Address == "" {
		return nil, fmt.Errorf("%w: address is required", ErrInvalidNode)
	}

	// Validate address format
	if !ValidateAddress(req.Address) {
		return nil, fmt.Errorf("%w: %s", ErrInvalidAddress, req.Address)
	}

	// Set defaults
	port := req.Port
	if port <= 0 {
		port = 8443
	}
	weight := req.Weight
	if weight <= 0 {
		weight = 1
	}

	// Generate token
	token, err := GenerateToken()
	if err != nil {
		s.logger.Error("Failed to generate token", logger.Err(err))
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	tagsJSON, _ := json.Marshal(req.Tags)
	ipWhitelistJSON, _ := json.Marshal(req.IPWhitelist)

	repoNode := &repository.Node{
		Name:        req.Name,
		Address:     req.Address,
		Port:        port,
		Token:       token,
		Status:      repository.NodeStatusOffline,
		Tags:        string(tagsJSON),
		Region:      req.Region,
		Weight:      weight,
		MaxUsers:    req.MaxUsers,
		SyncStatus:  repository.NodeSyncStatusPending,
		IPWhitelist: string(ipWhitelistJSON),
	}

	if err := s.nodeRepo.Create(ctx, repoNode); err != nil {
		s.logger.Error("Failed to create node", logger.Err(err))
		return nil, err
	}

	return s.toNode(repoNode), nil
}

// GetByID retrieves a node by ID.
func (s *Service) GetByID(ctx context.Context, id int64) (*Node, error) {
	repoNode, err := s.nodeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrNodeNotFound
	}
	return s.toNode(repoNode), nil
}

// Update updates a node.
func (s *Service) Update(ctx context.Context, id int64, req *UpdateNodeRequest) (*Node, error) {
	repoNode, err := s.nodeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrNodeNotFound
	}

	if req.Name != nil {
		repoNode.Name = *req.Name
	}
	if req.Address != nil {
		if !ValidateAddress(*req.Address) {
			return nil, fmt.Errorf("%w: %s", ErrInvalidAddress, *req.Address)
		}
		repoNode.Address = *req.Address
	}
	if req.Port != nil {
		repoNode.Port = *req.Port
	}
	if req.Tags != nil {
		tagsJSON, _ := json.Marshal(*req.Tags)
		repoNode.Tags = string(tagsJSON)
	}
	if req.Region != nil {
		repoNode.Region = *req.Region
	}
	if req.Weight != nil {
		repoNode.Weight = *req.Weight
	}
	if req.MaxUsers != nil {
		repoNode.MaxUsers = *req.MaxUsers
	}
	if req.IPWhitelist != nil {
		ipWhitelistJSON, _ := json.Marshal(*req.IPWhitelist)
		repoNode.IPWhitelist = string(ipWhitelistJSON)
	}

	if err := s.nodeRepo.Update(ctx, repoNode); err != nil {
		s.logger.Error("Failed to update node", logger.Err(err), logger.F("id", id))
		return nil, err
	}

	return s.toNode(repoNode), nil
}


// Delete deletes a node and reassigns its users to other healthy nodes.
func (s *Service) Delete(ctx context.Context, id int64) error {
	_, err := s.nodeRepo.GetByID(ctx, id)
	if err != nil {
		return ErrNodeNotFound
	}

	// Get users assigned to this node
	userIDs, err := s.assignmentRepo.GetUserIDsByNodeID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get users for node", logger.Err(err), logger.F("node_id", id))
		return err
	}

	// Reassign users to other healthy nodes
	if len(userIDs) > 0 {
		if err := s.reassignUsersFromNode(ctx, id, userIDs); err != nil {
			s.logger.Error("Failed to reassign users", logger.Err(err), logger.F("node_id", id))
			return err
		}
	}

	// Delete the node
	if err := s.nodeRepo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete node", logger.Err(err), logger.F("id", id))
		return err
	}

	s.logger.Info("Node deleted", logger.F("id", id), logger.F("reassigned_users", len(userIDs)))
	return nil
}

// reassignUsersFromNode reassigns users from a node to other healthy nodes.
func (s *Service) reassignUsersFromNode(ctx context.Context, excludeNodeID int64, userIDs []int64) error {
	// Get available healthy nodes
	healthyNodes, err := s.nodeRepo.GetAvailable(ctx)
	if err != nil {
		return err
	}

	// Filter out the node being deleted
	var availableNodes []*repository.Node
	for _, n := range healthyNodes {
		if n.ID != excludeNodeID {
			availableNodes = append(availableNodes, n)
		}
	}

	if len(availableNodes) == 0 {
		// No healthy nodes available, just delete assignments
		s.logger.Warn("No healthy nodes available for reassignment, deleting assignments",
			logger.F("node_id", excludeNodeID),
			logger.F("user_count", len(userIDs)))
		return s.assignmentRepo.DeleteByNodeID(ctx, excludeNodeID)
	}

	// Distribute users across available nodes using round-robin
	for i, userID := range userIDs {
		targetNode := availableNodes[i%len(availableNodes)]
		if err := s.assignmentRepo.Reassign(ctx, userID, targetNode.ID); err != nil {
			s.logger.Error("Failed to reassign user",
				logger.Err(err),
				logger.F("user_id", userID),
				logger.F("target_node_id", targetNode.ID))
			// Continue with other users
		}
	}

	return nil
}

// List lists nodes with filter and pagination.
func (s *Service) List(ctx context.Context, filter NodeFilter) ([]*Node, int64, error) {
	repoFilter := &repository.NodeFilter{
		Status:  filter.Status,
		Region:  filter.Region,
		Tags:    filter.Tags,
		GroupID: filter.GroupID,
		Limit:   filter.Limit,
		Offset:  filter.Offset,
	}

	repoNodes, err := s.nodeRepo.List(ctx, repoFilter)
	if err != nil {
		s.logger.Error("Failed to list nodes", logger.Err(err))
		return nil, 0, err
	}

	total, err := s.nodeRepo.Count(ctx, repoFilter)
	if err != nil {
		s.logger.Error("Failed to count nodes", logger.Err(err))
		return nil, 0, err
	}

	nodes := make([]*Node, len(repoNodes))
	for i, rn := range repoNodes {
		nodes[i] = s.toNode(rn)
	}

	return nodes, total, nil
}

// UpdateStatus updates a node's status.
func (s *Service) UpdateStatus(ctx context.Context, id int64, status string) error {
	if err := s.nodeRepo.UpdateStatus(ctx, id, status); err != nil {
		s.logger.Error("Failed to update node status", logger.Err(err), logger.F("id", id), logger.F("status", status))
		return err
	}
	return nil
}

// UpdateMetrics updates a node's metrics.
func (s *Service) UpdateMetrics(ctx context.Context, id int64, metrics *NodeMetrics) error {
	if err := s.nodeRepo.UpdateMetrics(ctx, id, metrics.Latency, metrics.CurrentUsers); err != nil {
		s.logger.Error("Failed to update node metrics", logger.Err(err), logger.F("id", id))
		return err
	}
	return nil
}

// UpdateLastSeen updates a node's last seen timestamp.
func (s *Service) UpdateLastSeen(ctx context.Context, id int64) error {
	if err := s.nodeRepo.UpdateLastSeen(ctx, id, time.Now()); err != nil {
		s.logger.Error("Failed to update node last seen", logger.Err(err), logger.F("id", id))
		return err
	}
	return nil
}

// GetHealthyNodes returns all healthy (online) nodes.
func (s *Service) GetHealthyNodes(ctx context.Context) ([]*Node, error) {
	repoNodes, err := s.nodeRepo.GetHealthy(ctx)
	if err != nil {
		return nil, err
	}

	nodes := make([]*Node, len(repoNodes))
	for i, rn := range repoNodes {
		nodes[i] = s.toNode(rn)
	}
	return nodes, nil
}

// GetAvailableNodes returns nodes that are online and not at capacity.
func (s *Service) GetAvailableNodes(ctx context.Context) ([]*Node, error) {
	repoNodes, err := s.nodeRepo.GetAvailable(ctx)
	if err != nil {
		return nil, err
	}

	nodes := make([]*Node, len(repoNodes))
	for i, rn := range repoNodes {
		nodes[i] = s.toNode(rn)
	}
	return nodes, nil
}


// ============================================
// Token Management
// ============================================

// GenerateToken generates a cryptographically secure random token.
func GenerateToken() (string, error) {
	bytes := make([]byte, TokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateNodeToken generates a new token for a node.
func (s *Service) GenerateNodeToken(ctx context.Context, nodeID int64) (string, error) {
	_, err := s.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return "", ErrNodeNotFound
	}

	token, err := GenerateToken()
	if err != nil {
		s.logger.Error("Failed to generate token", logger.Err(err))
		return "", err
	}

	if err := s.nodeRepo.UpdateToken(ctx, nodeID, token); err != nil {
		s.logger.Error("Failed to update node token", logger.Err(err), logger.F("node_id", nodeID))
		return "", err
	}

	s.logger.Info("Generated new token for node", logger.F("node_id", nodeID))
	return token, nil
}

// RotateToken rotates a node's token, invalidating the old one immediately.
func (s *Service) RotateToken(ctx context.Context, nodeID int64) (string, error) {
	node, err := s.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return "", ErrNodeNotFound
	}

	// Generate new token
	newToken, err := GenerateToken()
	if err != nil {
		s.logger.Error("Failed to generate new token", logger.Err(err))
		return "", err
	}

	// Update token (old token is immediately invalidated)
	if err := s.nodeRepo.UpdateToken(ctx, nodeID, newToken); err != nil {
		s.logger.Error("Failed to rotate token", logger.Err(err), logger.F("node_id", nodeID))
		return "", err
	}

	s.logger.Info("Rotated token for node",
		logger.F("node_id", nodeID),
		logger.F("node_name", node.Name))
	return newToken, nil
}

// RevokeToken revokes a node's token by setting it to empty.
func (s *Service) RevokeToken(ctx context.Context, nodeID int64) error {
	node, err := s.nodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return ErrNodeNotFound
	}

	// Set token to empty string to revoke it
	if err := s.nodeRepo.UpdateToken(ctx, nodeID, ""); err != nil {
		s.logger.Error("Failed to revoke token", logger.Err(err), logger.F("node_id", nodeID))
		return err
	}

	// Also set node status to offline since it can no longer authenticate
	if err := s.nodeRepo.UpdateStatus(ctx, nodeID, repository.NodeStatusOffline); err != nil {
		s.logger.Error("Failed to update node status after token revocation",
			logger.Err(err), logger.F("node_id", nodeID))
	}

	s.logger.Info("Revoked token for node",
		logger.F("node_id", nodeID),
		logger.F("node_name", node.Name))
	return nil
}

// ValidateToken validates a token and returns the associated node.
func (s *Service) ValidateToken(ctx context.Context, token string) (*Node, error) {
	if token == "" {
		return nil, ErrInvalidToken
	}

	repoNode, err := s.nodeRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Check if token is revoked (empty token in DB means revoked)
	if repoNode.Token == "" {
		return nil, ErrTokenRevoked
	}

	return s.toNode(repoNode), nil
}

// ============================================
// Address Validation
// ============================================

// domainRegex matches valid domain names
var domainRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

// ValidateAddress validates if the given address is a valid IPv4, IPv6, or domain name.
func ValidateAddress(address string) bool {
	if address == "" {
		return false
	}

	// Trim whitespace
	address = strings.TrimSpace(address)

	// Check if it's a valid IP address (IPv4 or IPv6)
	if ip := net.ParseIP(address); ip != nil {
		return true
	}

	// Check if it's a valid domain name
	if domainRegex.MatchString(address) {
		return true
	}

	// Check for localhost
	if address == "localhost" {
		return true
	}

	return false
}

// ValidateIPv4 validates if the given address is a valid IPv4 address.
func ValidateIPv4(address string) bool {
	ip := net.ParseIP(address)
	if ip == nil {
		return false
	}
	return ip.To4() != nil
}

// ValidateIPv6 validates if the given address is a valid IPv6 address.
func ValidateIPv6(address string) bool {
	ip := net.ParseIP(address)
	if ip == nil {
		return false
	}
	return ip.To4() == nil && ip.To16() != nil
}

// ValidateDomain validates if the given address is a valid domain name.
func ValidateDomain(address string) bool {
	if address == "" {
		return false
	}
	address = strings.TrimSpace(address)
	if address == "localhost" {
		return true
	}
	return domainRegex.MatchString(address)
}


// ============================================
// Helper Methods
// ============================================

// toNode converts a repository node to a service node.
func (s *Service) toNode(rn *repository.Node) *Node {
	var tags []string
	var ipWhitelist []string

	if rn.Tags != "" {
		_ = json.Unmarshal([]byte(rn.Tags), &tags)
	}
	if rn.IPWhitelist != "" {
		_ = json.Unmarshal([]byte(rn.IPWhitelist), &ipWhitelist)
	}

	return &Node{
		ID:           rn.ID,
		Name:         rn.Name,
		Address:      rn.Address,
		Port:         rn.Port,
		Token:        rn.Token,
		Status:       rn.Status,
		Tags:         tags,
		Region:       rn.Region,
		Weight:       rn.Weight,
		MaxUsers:     rn.MaxUsers,
		CurrentUsers: rn.CurrentUsers,
		Latency:      rn.Latency,
		LastSeenAt:   rn.LastSeenAt,
		SyncStatus:   rn.SyncStatus,
		SyncedAt:     rn.SyncedAt,
		IPWhitelist:  ipWhitelist,
		CreatedAt:    rn.CreatedAt,
		UpdatedAt:    rn.UpdatedAt,
	}
}

// GetStatistics returns node statistics.
func (s *Service) GetStatistics(ctx context.Context) (map[string]int64, error) {
	return s.nodeRepo.CountByStatus(ctx)
}

// GetTotalUsers returns the total number of users across all nodes.
func (s *Service) GetTotalUsers(ctx context.Context) (int64, error) {
	return s.nodeRepo.GetTotalUsers(ctx)
}
