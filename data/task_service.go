package data

// imports
import (
	"context";
	"errors";
	"fmt";
	"log";
	"time";
	"github.com/natnael-eyuel-dev/Task-Management-API-with-JWT-Auth/models";
	"go.mongodb.org/mongo-driver/bson";
	"go.mongodb.org/mongo-driver/bson/primitive";
	"go.mongodb.org/mongo-driver/mongo";
	"go.mongodb.org/mongo-driver/mongo/options";
)

type TaskManager interface {
	CreateTask(task *models.Task) (*models.Task, error)                     // create new task with validation
	DeleteTask(taskID string) error                 			// delete existing task or return error if not found
	GetAllTasks() ([]models.Task, error)         				// get all tasks in the system
	GetTaskByID(taskID string) (*models.Task, error) 		        // get specific task by id or return error if not found
	UpdateTask(taskID string, task *models.Task) (*models.Task, error)      // update existing task or return error if not found
}

type MongoDBTaskManager struct {
	client     	 *mongo.Client      // connection to mongodb
	database         string             // which database to use
	collection       string             // which collection to work with
}

// create a new connection to mongodb
func NewMongoDBTaskManager(uri, db, colln string) (*MongoDBTaskManager, error) {
	
	clientOptions := options.Client().ApplyURI(uri)    // set client options

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)    // set timeout
	defer cancel()
	 
	client, err := mongo.Connect(ctx, clientOptions)      // trying to connect with error handling 
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)     // check the connection
	if err != nil {
		return nil, err
	}

	log.Println("Connected to MongoDB!")
	return &MongoDBTaskManager{
		client:     client,
		database:   db,
		collection: colln,
	}, nil
}

// add new task to database 
func (taskServ *MongoDBTaskManager) collectionRef() *mongo.Collection {
	return taskServ.client.Database(taskServ.database).Collection(taskServ.collection)
}

func (taskServ *MongoDBTaskManager) CreateTask(task *models.Task) (*models.Task, error) {

	// validate task fields before creation
	if task.Title == "" {
		return nil, errors.New("task title can not be empty")
	}
	if task.Description == "" {
		return nil, errors.New("task description can not be empty")
	}
	if task.DueDate.IsZero() {
		return nil, errors.New("task duedate can not be empty")
	}
	if task.Status == "" {
		return nil, errors.New("task status can not be empty")
	}

	collection := taskServ.collectionRef()

	contx, cancel := context.WithTimeout(context.Background(), 5*time.Second)     // set timeout
	defer cancel()

	task.ID = primitive.NewObjectID()               // create a unique id for the new task
	_, err := collection.InsertOne(contx, task)     // create the new task with error handling
	if err != nil {
        return nil, fmt.Errorf("failed to create task: %v", err)
    }

	return task, nil       // return the new created task and nil
}

// remove a task from the database 
func (taskServ *MongoDBTaskManager) DeleteTask(taskID string) error {
	
	var task models.Task
	collection := taskServ.collectionRef()

	objID, err := primitive.ObjectIDFromHex(taskID)       // convert string id to mongodb's id format with error handling 
	if err != nil {
		return err
	}

	contx, cancel := context.WithTimeout(context.Background(), 5*time.Second)        // set timeout
	defer cancel()

	err = collection.FindOne(contx, bson.M{"_id": objID}).Decode(&task)        // check task with the id in the database
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return errors.New("no task found with this id to delete")
        }
        return err
    }

	_, err = collection.DeleteOne(contx, bson.M{"_id":objID})         // delete the task with error handling
	if err != nil {
		return err
	}

	return nil       // return nil
}

func (taskServ *MongoDBTaskManager) GetAllTasks() ([]models.Task, error) {
	
	var allTasks []models.Task
	collection := taskServ.collectionRef()

	contx, cancel := context.WithTimeout(context.Background(), 5*time.Second)      // set timeout
	defer cancel()

	cursor, err := collection.Find(contx, bson.M{})      // find all documents in the collection
	if err != nil {
		return nil, err
	}
	
	defer cursor.Close(contx)      // close cursor when done

	err = cursor.All(contx, &allTasks)       // read all result into our slice
	if err != nil {
		return nil, err
	}

	return allTasks, nil     // return all tasks in the database and nil
}

// find one specific task by its id
func (taskServ *MongoDBTaskManager) GetTaskByID(taskID string) (*models.Task, error) {
	
	var task models.Task
	collection := taskServ.collectionRef()

	objID, err := primitive.ObjectIDFromHex(taskID)      // convert string id to mongodb's format with error handling 
	if err != nil {
		return nil, err
	}

	contx, cancel := context.WithTimeout(context.Background(), 5*time.Second)      // set timeout
	defer cancel()

	err = collection.FindOne(contx, bson.M{"_id":objID}).Decode(&task)      // check if task exists
	if err != nil {
		return nil, errors.New("no task found with this id to see")
	}

	return &task, nil    // return the found task and nil
}

// update an existing task's details
func (taskServ *MongoDBTaskManager) UpdateTask(taskID string, taskUpdate *models.Task) (*models.Task, error) {
	
	var updatedtask models.Task
	collection := taskServ.collectionRef()

	objID, err := primitive.ObjectIDFromHex(taskID)      // convert string id to mongodb's format with error handling 
	if err != nil {
		return nil, err
	}

	contx, cancel := context.WithTimeout(context.Background(), 5*time.Second)     // set timeout
	defer cancel()

	err = collection.FindOne(contx, bson.M{"_id": objID}).Decode(&updatedtask)     // check if task exists
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, errors.New("no task found with this id to update")
        }
        return nil, err
    }

	update := bson.M{"$set": bson.M{}}        // prepare what we want to change
	setFields := update["$set"].(bson.M)

	// only update fields that were actually provided
	if taskUpdate.Title != "" {
        setFields["title"] = taskUpdate.Title
    }
    if taskUpdate.Description != "" {
        setFields["description"] = taskUpdate.Description
    }
    if !taskUpdate.DueDate.IsZero() {
        setFields["due_date"] = taskUpdate.DueDate
    }
    if taskUpdate.Status != "" {
        setFields["status"] = taskUpdate.Status
    }

	// stop if nothing valid to update
	if len(setFields) == 0 {
        return nil, errors.New("no valid fields provided for update")
    }

	opts := options.FindOneAndUpdate().        // to get updated document back 
		SetReturnDocument(options.After)

	// perform update and get the updated task
	err = collection.FindOneAndUpdate(
		contx, 
		bson.M{"_id": objID},
		update,
		opts,
	).Decode(&updatedtask)

	if err != nil {
		return nil, err
	}

	return &updatedtask, nil  // return the updated task and nil
}

// close mongodb connection
func (taskServ *MongoDBTaskManager) Close() error {
	contx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return taskServ.client.Disconnect(contx)
}
