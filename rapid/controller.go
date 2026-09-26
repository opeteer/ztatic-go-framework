package rapid

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/labstack/echo/v5"
)

// Resource defines a standard generic RESTful interface for any data model.
type Resource[T any] interface {
	FindAll(c *echo.Context) ([]T, error)
	FindByID(c *echo.Context, id string) (T, error)
	Create(c *echo.Context, item *T) (T, error)
	Update(c *echo.Context, id string, item *T) (T, error)
	Delete(c *echo.Context, id string) error
}

// RegisterResource is a rapid DX scaffolder. It automatically mounts standard 
// RESTful endpoints (GET, POST, PUT, DELETE) onto an Echo Group for the given 
// generic Resource interface. It automatically integrates Ztatic's `BindAndValidate`.
func RegisterResource[T any](g *echo.Group, path string, res Resource[T]) {
	var group *echo.Group
	cleanPath := strings.Trim(path, "/")
	if cleanPath == "" {
		group = g
	} else {
		group = g.Group("/" + cleanPath)
	}
	
	// Introspect type T for OpenAPI docs
	var item T
	itemType := reflect.TypeOf(item)
	for itemType != nil && itemType.Kind() == reflect.Pointer {
		itemType = itemType.Elem()
	}
	typeName := "Resource"
	if itemType != nil {
		typeName = itemType.Name()
	}

	// GET /resource - FindAll
	rFindAll := group.GET("", func(c *echo.Context) error {
		items, err := res.FindAll(c)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, items)
	})
	DefaultOpenAPIGenerator.RegisterRouteMeta(rFindAll.Method, rFindAll.Path, RouteMetadata{
		Summary:      "Find all " + typeName + "s",
		Tags:         []string{typeName},
		ResponseType: reflect.SliceOf(itemType),
	})

	// GET /resource/:id - FindByID
	rFindByID := group.GET("/:id", func(c *echo.Context) error {
		id := c.Param("id")
		item, err := res.FindByID(c, id)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, item)
	})
	DefaultOpenAPIGenerator.RegisterRouteMeta(rFindByID.Method, rFindByID.Path, RouteMetadata{
		Summary:      "Find " + typeName + " by ID",
		Tags:         []string{typeName},
		ResponseType: itemType,
	})

	// POST /resource - Create
	rCreate := group.POST("", func(c *echo.Context) error {
		var item T
		// Ztatic DX: Automatically binds the payload and executes struct validation tags.
		if err := BindAndValidate(c, &item); err != nil {
			return err
		}
		
		created, err := res.Create(c, &item)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusCreated, created)
	})
	DefaultOpenAPIGenerator.RegisterRouteMeta(rCreate.Method, rCreate.Path, RouteMetadata{
		Summary:      "Create " + typeName,
		Tags:         []string{typeName},
		RequestType:  itemType,
		ResponseType: itemType,
	})

	// PUT /resource/:id - Update
	rUpdate := group.PUT("/:id", func(c *echo.Context) error {
		id := c.Param("id")
		var item T
		
		if err := BindAndValidate(c, &item); err != nil {
			return err
		}

		updated, err := res.Update(c, id, &item)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, updated)
	})
	DefaultOpenAPIGenerator.RegisterRouteMeta(rUpdate.Method, rUpdate.Path, RouteMetadata{
		Summary:      "Update " + typeName,
		Tags:         []string{typeName},
		RequestType:  itemType,
		ResponseType: itemType,
	})

	// DELETE /resource/:id - Delete
	rDelete := group.DELETE("/:id", func(c *echo.Context) error {
		id := c.Param("id")
		if err := res.Delete(c, id); err != nil {
			return err
		}
		return c.NoContent(http.StatusNoContent)
	})
	DefaultOpenAPIGenerator.RegisterRouteMeta(rDelete.Method, rDelete.Path, RouteMetadata{
		Summary: "Delete " + typeName,
		Tags:    []string{typeName},
	})
}
