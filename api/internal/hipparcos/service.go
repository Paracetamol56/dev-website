package hipparcos

import (
	"context"
	"dev/internal/utils"
)

func ListStars(ctx context.Context) ([]HipparcosHR, error) {
	return GetAllHipparcosHR(ctx)
}

func GetStar(ctx context.Context, hip int) (*HipparcosHR, error) {
	stars, err := GetHipparcosHRByHIP(ctx, hip)
	if err != nil {
		return nil, err
	}
	if len(stars) == 0 {
		return nil, utils.NotFound("Not found")
	}
	return &stars[0], nil
}
