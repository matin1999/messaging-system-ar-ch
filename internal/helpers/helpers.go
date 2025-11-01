package helpers

import (
	"errors"
	"fmt"
	"postchi/pkg/db"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func ParseUintParam(c *fiber.Ctx, name string) (uint, error) {
	raw := c.Params(name)
	n, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s", name)
	}
	return uint(n), nil
}

func ToServiceType(s string) (db.ServiceType, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "express":
		return db.ServiceType("express"), nil
	case "indirect":
		return db.ServiceType("indirect"), nil
	default:
		return "", errors.New("type must be 'express' or 'indirect'")
	}
}
