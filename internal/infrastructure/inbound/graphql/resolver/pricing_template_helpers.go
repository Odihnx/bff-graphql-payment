package resolver

import (
	"bff-graphql-payment/graph/model"
	domainmodel "bff-graphql-payment/internal/domain/model"
	"context"
	"fmt"
	"strings"
)

func pricingCellsFromInput(cells []*model.PricingCellInput) []domainmodel.PricingCell {
	result := make([]domainmodel.PricingCell, 0, len(cells))
	for _, c := range cells {
		if c == nil {
			continue
		}
		result = append(result, domainmodel.PricingCell{
			BookingTimeID: c.BookingTimeID,
			Size:          strings.ToLower(string(c.Size)),
			Price:         c.Price,
		})
	}
	return result
}

func (r *mutationResolver) resolvePricingTemplate(ctx context.Context, input model.CreateRackPaymentInput) (int, []domainmodel.GroupProductBySize, error) {
	cells := pricingCellsFromInput(input.PricingCells)

	match, err := r.paymentInfraService.FindExactPricingTemplate(ctx, cells)
	if err != nil {
		return 0, nil, err
	}
	if match != nil {
		return match.PricingTemplateID, match.GroupProductBySize, nil
	}

	installationName := ""
	if input.InstallationName != nil {
		installationName = *input.InstallationName
	}
	rackDescription := ""
	if input.RackDescription != nil {
		rackDescription = *input.RackDescription
	}
	name := strings.TrimSpace(installationName + " · " + rackDescription)
	if name == "·" || name == "" {
		name = fmt.Sprintf("Rack %d", input.RackReference)
	}

	created, err := r.paymentInfraService.CreatePricingTemplate(ctx, name, "Plantilla generada automáticamente para este rack", cells)
	if err != nil {
		return 0, nil, err
	}
	return created.PricingTemplateID, created.GroupProductBySize, nil
}
