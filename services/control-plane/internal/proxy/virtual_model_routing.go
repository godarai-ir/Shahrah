package proxy

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/virtualmodels"
)

type virtualModelRoutePlan struct {
	model        *platform.VirtualModel
	targets      []platform.VirtualModelTarget
	fromSnapshot bool
}

func resolveVirtualModelRoutePlan(
	ctx context.Context,
	authInfo *platform.AuthContext,
	resolver virtualmodels.VirtualModelResolver,
	engine virtualmodels.RoutingEngine,
	tenantID string,
	requestedModel string,
) (*virtualModelRoutePlan, error) {
	if engine == nil {
		return nil, virtualmodels.ErrNotVirtualModel
	}

	virtualModel, fromSnapshot := buildVirtualModelFromSnapshot(authInfo, requestedModel)
	if !fromSnapshot {
		if resolver == nil {
			return nil, virtualmodels.ErrNotVirtualModel
		}

		var err error

		virtualModel, err = resolver.Resolve(ctx, tenantID, requestedModel)
		if err != nil {
			return nil, err
		}
	}

	targets, err := engine.SelectTargets(virtualModel)
	if err != nil {
		return nil, err
	}

	if len(targets) == 0 {
		return nil, virtualmodels.ErrNoEligibleTargets
	}

	return &virtualModelRoutePlan{
		model:        virtualModel,
		targets:      targets,
		fromSnapshot: fromSnapshot,
	}, nil
}

func resolveFirstVirtualDeployment(
	ctx context.Context,
	tenantID string,
	targets []platform.VirtualModelTarget,
	llmRouter *router.Router,
) *router.Deployment {
	for targetIndex := range targets {
		deployment, err := resolveVirtualDeploymentStrict(
			ctx,
			llmRouter,
			tenantID,
			deploymentFromVirtualTarget(targets[targetIndex]),
		)
		if err == nil {
			return deployment
		}
	}

	return nil
}

func resolveVirtualDeploymentStrict(
	ctx context.Context,
	llmRouter *router.Router,
	tenantID string,
	deployment *router.Deployment,
) (*router.Deployment, error) {
	if llmRouter == nil || deployment == nil {
		return nil, errors.New("virtual model deployment router is unavailable")
	}

	modelName := strings.TrimSpace(deployment.ModelName)
	if modelName == "" {
		return nil, errors.New("virtual model target is missing a model identifier")
	}

	targetTimeout := deployment.Params.Timeout

	resolved, err := llmRouter.Route(ctx, tenantID, modelName)
	if err != nil {
		return nil, fmt.Errorf("resolve virtual model target deployment: %w", err)
	}

	if resolved == nil {
		return nil, errors.New("resolve virtual model target deployment: empty deployment")
	}

	out := *resolved
	if targetTimeout > 0 {
		out.Params.Timeout = targetTimeout
	}

	return &out, nil
}

func cloneStringMap(input map[string]string) map[string]string {
	if input == nil {
		return map[string]string{}
	}

	out := make(map[string]string, len(input))
	for key, value := range input {
		out[key] = value
	}

	return out
}

func deleteHeaderFold(headers map[string]string, name string) {
	for key := range headers {
		if strings.EqualFold(strings.TrimSpace(key), name) {
			delete(headers, key)
		}
	}
}

func replaceDeploymentRequestMetadata(
	input map[string]interface{},
	deployment platform.Deployment,
) map[string]interface{} {
	out := make(map[string]interface{}, len(input))
	for key, value := range input {
		out[key] = value
	}

	for _, key := range []string{
		"credential_source",
		"billing_mode",
		"provider_account_id",
		"connection_id",
		"credential_version_id",
		"provider_account_name",
	} {
		delete(out, key)
	}

	return enrichRequestMetadataForDeployment(out, deployment)
}
