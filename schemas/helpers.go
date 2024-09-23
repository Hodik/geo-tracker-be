package schemas

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

func ValidatePolygonWKT(polygon string) error {
	if !strings.HasPrefix(polygon, "POLYGON((") || !strings.HasSuffix(polygon, "))") {
		return errors.New("invalid polygon format: must start with 'POLYGON((' and end with '))'")
	}

	coordPart := strings.TrimPrefix(polygon, "POLYGON((")
	coordPart = strings.TrimSuffix(coordPart, "))")

	points := strings.Split(coordPart, ",")

	if len(points) < 4 {
		return errors.New("invalid polygon: must have at least 4 points")
	}

	if strings.TrimSpace(points[0]) != strings.TrimSpace(points[len(points)-1]) {
		return errors.New("invalid polygon: first and last points must be the same (closed polygon)")
	}

	coordRegex := regexp.MustCompile(`^[-+]?[0-9]*\.?[0-9]+ [-+]?[0-9]*\.?[0-9]+$`)
	for _, point := range points {
		point = strings.TrimSpace(point)
		if !coordRegex.MatchString(point) {
			return fmt.Errorf("invalid coordinate format: '%s'", point)
		}
	}

	return nil
}
