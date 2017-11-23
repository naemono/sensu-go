package eventd

import (
	"context"
	"fmt"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// getProxyEntity verifies if a source was provided in the given event and if
// so, retrieves the corresponding entity in the store in order to replace the
// event's entity with it. In case no entity exists, we create an entity with
// the proxy class
func getProxyEntity(ctx context.Context, event *types.Event, s store.Store) error {
	// Verify if a source, representing a proxy entity, is defined in the check
	if event.Check.Config.Source != "" {
		// Query the store for an entity using the given source as the entity ID
		entity, err := s.GetEntityByID(ctx, event.Check.Config.Source)
		if err != nil {
			return fmt.Errorf("could not query the store for a proxy entity: %s", err.Error())
		}

		// Check if an entity was found for this source. If not, we need to create it
		if entity == nil {
			entity = &types.Entity{
				ID:           event.Check.Config.Source,
				Class:        types.EntityProxyClass,
				Environment:  event.Entity.Environment,
				Organization: event.Entity.Organization,
			}

			if err := s.UpdateEntity(ctx, entity); err != nil {
				return fmt.Errorf("could not create a proxy entity: %s", err.Error())
			}
		}

		event.Entity = entity
	}

	return nil
}
