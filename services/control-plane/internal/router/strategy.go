package router

// RoutingStrategy defines the interface for routing strategies.
type RoutingStrategy interface {
	Name() string
	SelectDeployment(deployments []*Deployment, model string) (*Deployment, error)
}
