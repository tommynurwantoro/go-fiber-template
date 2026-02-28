package handler

import (
	"app/internal/application/model"
	"app/internal/application/service"

	"github.com/gofiber/fiber/v2"
)

type HealthCheckHandler interface {
	Check(c *fiber.Ctx) error
}

type HealthCheckHandlerImpl struct {
	HealthCheckService service.HealthCheckService `inject:"healthCheckService"`
}

func (h *HealthCheckHandlerImpl) addServiceStatus(
	serviceList *[]model.HealthCheck, name string, isUp bool, message *string,
) {
	status := "Up"

	if !isUp {
		status = "Down"
	}

	*serviceList = append(*serviceList, model.HealthCheck{
		Name:    name,
		Status:  status,
		IsUp:    isUp,
		Message: message,
	})
}

// @Tags         Health
// @Summary      Health check
// @Description  Check the status of services (PostgreSQL, Memory). Returns 200 when all healthy, 500 when any service is down.
// @Accept       json
// @Produce      json
// @Router       /health-check [get]
// @Success      200  {object}  model.HealthCheckResponse  "All services healthy"
// @Failure      500  {object}  model.HealthCheckAPIResponseError  "One or more services unhealthy"
func (h *HealthCheckHandlerImpl) Check(c *fiber.Ctx) error {
	isHealthy := true
	var serviceList []model.HealthCheck

	// Check the database connection
	if err := h.HealthCheckService.GormCheck(); err != nil {
		isHealthy = false
		errMsg := err.Error()
		h.addServiceStatus(&serviceList, "Postgre", false, &errMsg)
	} else {
		h.addServiceStatus(&serviceList, "Postgre", true, nil)
	}

	if err := h.HealthCheckService.MemoryHeapCheck(); err != nil {
		isHealthy = false
		errMsg := err.Error()
		h.addServiceStatus(&serviceList, "Memory", false, &errMsg)
	} else {
		h.addServiceStatus(&serviceList, "Memory", true, nil)
	}

	// Return the response based on health check result
	statusCode := fiber.StatusOK
	status := "success"

	if !isHealthy {
		statusCode = fiber.StatusInternalServerError
		status = "error"
	}

	return c.Status(statusCode).JSON(model.HealthCheckResponse{
		Code:      statusCode,
		Status:    status,
		Message:   "Health check completed",
		IsHealthy: isHealthy,
		Result:    serviceList,
	})
}
