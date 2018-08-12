
package main

import (
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// ToDo type
type ToDo struct {
	Title     string `bson:"title"`
	Completed bool   `bson:"completed"`
}

func main() {
	app := iris.New()
	/**
	* Logging
	 */
	app.Logger().SetLevel("debug")
	// Recover from panics and log the panic message to the application's logger ("Warn" level).
	app.Use(recover.New())
	// logs HTTP requests to the application's logger ("Info" level)
	app.Use(logger.New())

	/**
	* Mongo
	 */

	// Connection variables
	const (
		Host       = ""
		Username   = ""
		Password   = ""
		Database   = ""
		Collection = ""
	)

	// Mongo connection
	session, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{Host},
		Username: Username,
		Password: Password,
		Database: Database,
	})
	// If there is an error connecting to Mongo - panic
	if err != nil {
		panic(err)
	}
	// Close session when surrounding function has returned/ended
	defer session.Close()
	// mongogo is the database name
	db := session.DB(Database)
	// todos is the collection name
	collection := db.C(Collection)

	/**
	* Routes
	 */

	// Create todo using POST Request body
	app.Post("/addOne", func(ctx iris.Context) {
		// Create a new ToDo
		var todo ToDo
		// Pass the pointer of todo so it is updated with the result
		// which is the POST data
		err := ctx.ReadJSON(&todo)
		// If there is an error or no Title in the POST Body
		if err != nil || todo.Title == "" {
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.JSON(iris.Map{"message": "Post body must be a JSON object with at least a Title!"})
			return
		}
		// Insert into the database
		collection.Insert(todo)
	})

	// Get todo by Title
	app.Get("/title/{title:string}", func(ctx iris.Context) {
		title := ctx.Params().Get("title")
		var results []ToDo
		err := collection.Find(bson.M{"title": title}).All(&results)
		if err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.JSON(iris.Map{"message": "An error occured", "error": err})
			return
		}
		ctx.JSON(iris.Map{"results": results})
	})
	// Get todo by Completed status
	app.Get("/completed/{completed:boolean}", func(ctx iris.Context) {
		completed, _ := ctx.Params().GetBool("completed")
		var results []ToDo
		err := collection.Find(bson.M{"completed": completed}).All(&results)
		if err != nil {
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.JSON(iris.Map{"message": "An error occured", "error": err})
			return
		}
		ctx.JSON(iris.Map{"results": results})
	})
	// Update one todo by title
	app.Put("/title/{title:string}", func(ctx iris.Context) {
		// Get the title from the URL Parameter
		title := ctx.Params().Get("title")
		// Construct the query object
		query := bson.M{"title": title}
		// Create a new ToDo
		var change ToDo
		// Update change with the PUT body
		ctx.ReadJSON(&change)
		// Only update one record
		collection.Update(query, change)
	})
	// Update todo by completed
	app.Put("/completed/{completed:boolean}", func(ctx iris.Context) {
		// Get the completed status from the URL parameter
		completed, _ := ctx.Params().GetBool("completed")
		query := bson.M{"completed": completed}
		var change ToDo
		ctx.ReadJSON(&change)
		// Update all records
		collection.UpdateAll(query, bson.M{"$set": bson.M{"completed": change.Completed}})
	})
	// Remove by Title
	app.Delete("/title/{title:string}", func(ctx iris.Context) {
		title := ctx.Params().Get("title")
		collection.Remove(bson.M{"title": title})
	})
	// Remove by completed
	app.Delete("/completed/{completed:boolean}", func(ctx iris.Context) {
		completed, _ := ctx.Params().GetBool("completed")
		collection.RemoveAll(bson.M{"completed": completed})
	})

	// Run app on port 8080
	// ignore server closed errors ([ERRO] 2018/04/09 12:25 http: Server closed)
	app.Run(iris.Addr(":10086"), iris.WithoutServerError(iris.ErrServerClosed))
}

