package handler

import (
	"app/internal/application/model"
	"app/internal/application/service"
	"app/internal/pkg/formatter"
	"math"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/tommynurwantoro/golog"
)

type UserHandler interface {
	GetUsers(c *fiber.Ctx) error
	GetUserByID(c *fiber.Ctx) error
	CreateUser(c *fiber.Ctx) error
	UpdateUser(c *fiber.Ctx) error
	DeleteUser(c *fiber.Ctx) error
}

type UserHandlerImpl struct {
	UserService  service.UserService  `inject:"userService"`
	TokenService service.TokenService `inject:"tokenService"`
}

// @Tags         Users
// @Summary      Get all users
// @Description  Retrieve paginated list of users. Only admins (getUsers permission) can access. Supports search by name, email, or role.
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        page    query     int     false  "Page number"  default(1)
// @Param        limit   query     int     false  "Items per page"  default(10)
// @Param        search  query     string  false  "Search by name, email, or role"
// @Router       /v1/users [get]
// @Success      200  {object}  formatter.SuccessResponse{data=[]model.GetUserResponse,metadata=formatter.Metadata}
// @Failure      400  {object}  model.ErrorInvalidRequest  "Invalid query parameters"
// @Failure      401  {object}  model.ErrorUnauthorized  "Invalid or missing access token"
// @Failure      403  {object}  model.ErrorForbidden  "Insufficient permissions"
func (u *UserHandlerImpl) GetUsers(c *fiber.Ctx) error {
	query := &model.GetUserRequest{
		Page:   c.QueryInt("page", 1),
		Limit:  c.QueryInt("limit", 10),
		Search: c.Query("search", ""),
	}

	users, totalResults, err := u.UserService.GetUsers(c, query)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).
		JSON(formatter.NewSuccessResponseWithMetadata(formatter.Success, "Get all users successfully", users, formatter.Metadata{
			Page:         query.Page,
			Limit:        query.Limit,
			TotalPages:   int64(math.Ceil(float64(totalResults) / float64(query.Limit))),
			TotalResults: totalResults,
		}))
}

// @Tags         Users
// @Summary      Get a user
// @Description  Fetch user by ID. Users can fetch only their own data; admins (getUsers) can fetch any user.
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        userId  path  string  true  "User UUID"
// @Router       /v1/users/{userId} [get]
// @Success      200  {object}  model.GetUserResponse
// @Failure      400  {object}  model.ErrorInvalidRequest  "Invalid user ID format"
// @Failure      401  {object}  model.ErrorUnauthorized  "Invalid or missing access token"
// @Failure      403  {object}  model.ErrorForbidden  "Insufficient permissions"
// @Failure      404  {object}  model.ErrorNotFound  "User not found"
func (u *UserHandlerImpl) GetUserByID(c *fiber.Ctx) error {
	userID := c.Params("userId")

	if _, err := uuid.Parse(userID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID")
	}

	user, err := u.UserService.GetUserByID(c, userID)
	if err != nil {
		return err
	}

	resp := &model.GetUserResponse{
		ID:              user.ID.String(),
		Name:            user.Name,
		Email:           user.Email,
		Role:            user.Role,
		IsEmailVerified: user.VerifiedEmail,
	}

	return c.Status(fiber.StatusOK).
		JSON(formatter.NewSuccessResponse(formatter.Success, "Get user successfully", resp))
}

// @Tags         Users
// @Summary      Create a user
// @Description  Create a new user. Only admins (manageUsers permission) can create users.
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body  model.CreateUserRequest  true  "Request body (name, email, password, role: user|admin)"
// @Router       /v1/users [post]
// @Success      201  {object}  model.CreateUserResponse
// @Failure      400  {object}  model.ErrorInvalidRequest  "Invalid request body or validation failed"
// @Failure      401  {object}  model.ErrorUnauthorized  "Invalid or missing access token"
// @Failure      403  {object}  model.ErrorForbidden  "Insufficient permissions"
// @Failure      409  {object}  model.ErrorDuplicateEmail  "Email already in use"
func (u *UserHandlerImpl) CreateUser(c *fiber.Ctx) error {
	req := new(model.CreateUserRequest)

	if err := c.BodyParser(&req); err != nil {
		golog.Error("Error parsing request body", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	user, err := u.UserService.CreateUser(c, req)
	if err != nil {
		return err
	}

	resp := &model.CreateUserResponse{
		ID:              user.ID.String(),
		Name:            user.Name,
		Email:           user.Email,
		Role:            user.Role,
		IsEmailVerified: user.VerifiedEmail,
	}

	return c.Status(fiber.StatusCreated).
		JSON(formatter.NewSuccessResponse(formatter.Success, "Create user successfully", resp))
}

// @Tags         Users
// @Summary      Update a user
// @Description  Update user by ID. Users can update only their own data; admins (manageUsers) can update any user. Provide at least one of name, email, or password.
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        userId   path   string  true  "User UUID"
// @Param        request  body  model.UpdateUserRequest  true  "Request body (name, email, password - all optional)"
// @Router       /v1/users/{userId} [patch]
// @Success      200  {object}  model.UpdateUserResponse
// @Failure      400  {object}  model.ErrorInvalidRequest  "Invalid user ID or request body"
// @Failure      401  {object}  model.ErrorUnauthorized  "Invalid or missing access token"
// @Failure      403  {object}  model.ErrorForbidden  "Insufficient permissions"
// @Failure      404  {object}  model.ErrorNotFound  "User not found"
// @Failure      409  {object}  model.ErrorDuplicateEmail  "Email already in use"
func (u *UserHandlerImpl) UpdateUser(c *fiber.Ctx) error {
	req := new(model.UpdateUserRequest)
	userID := c.Params("userId")

	if _, err := uuid.Parse(userID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID")
	}

	req.UserID = userID

	if err := c.BodyParser(&req); err != nil {
		golog.Error("Error parsing request body", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	user, err := u.UserService.UpdateUser(c, req)
	if err != nil {
		return err
	}

	resp := &model.UpdateUserResponse{
		ID:              user.ID.String(),
		Name:            user.Name,
		Email:           user.Email,
		Role:            user.Role,
		IsEmailVerified: user.VerifiedEmail,
	}

	return c.Status(fiber.StatusOK).
		JSON(formatter.NewSuccessResponse(formatter.Success, "Update user successfully", resp))
}

// @Tags         Users
// @Summary      Delete a user
// @Description  Delete user by ID. Users can delete only themselves; admins (manageUsers) can delete any user. All tokens are revoked.
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        userId  path  string  true  "User UUID"
// @Router       /v1/users/{userId} [delete]
// @Success      200  {object}  model.SuccessMessageAPIResponse
// @Failure      400  {object}  model.ErrorInvalidRequest  "Invalid user ID format"
// @Failure      401  {object}  model.ErrorUnauthorized  "Invalid or missing access token"
// @Failure      403  {object}  model.ErrorForbidden  "Insufficient permissions"
// @Failure      404  {object}  model.ErrorNotFound  "User not found"
func (u *UserHandlerImpl) DeleteUser(c *fiber.Ctx) error {
	userID := c.Params("userId")

	if _, err := uuid.Parse(userID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID")
	}

	if err := u.TokenService.DeleteAllToken(c, userID); err != nil {
		return err
	}

	if err := u.UserService.DeleteUser(c, userID); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).
		JSON(formatter.NewSuccessResponse(formatter.Success, "Delete user successfully", nil))
}
