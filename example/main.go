package main

// Thanks [echo](https://github.com/labstack/echo) web framework for this useful example of CRUD API
import (
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type User struct {
	ID   int
	Name string
}

var (
	users = map[int]*User{
		1: &User{1, "First User"},
		2: &User{2, "Second User"},
	}
	seq = 3
)

//----------
// Handlers
//----------

func createUser() echo.HandlerFunc {
	return func(c echo.Context) error {
		u := &User{}
		if err := c.Bind(u); err != nil {
			return err
		}
		u.ID = seq
		users[u.ID] = u
		seq++
		return c.JSON(http.StatusCreated, u)
	}
}

func getUser() echo.HandlerFunc {
	return func(c echo.Context) error {
		id, _ := strconv.Atoi(c.Param("id"))
		user, ok := users[id]
		if !ok {
			return c.NoContent(http.StatusNotFound)
		}
		return c.JSON(http.StatusOK, user)
	}
}

func updateUser() echo.HandlerFunc {
	return func(c echo.Context) error {
		u := new(User)
		if err := c.Bind(u); err != nil {
			return err
		}
		id, _ := strconv.Atoi(c.Param("id"))
		users[id].Name = u.Name
		return c.JSON(http.StatusOK, users[id])
	}
}
func deleteUser() echo.HandlerFunc {
	return func(c echo.Context) error {
		id, _ := strconv.Atoi(c.Param("id"))
		delete(users, id)
		return c.NoContent(http.StatusNoContent)
	}
}

func main() {
	router := echo.New()

	// Middleware
	router.Use(middleware.Recover())
	router.Use(middleware.Logger())

	// Routes
	router.POST("/users", createUser())
	router.GET("/users/:id", getUser())
	router.PATCH("/users/:id", updateUser())
	router.DELETE("/users/:id", deleteUser())

	// Start server
	router.Start("localhost:1323")
}
