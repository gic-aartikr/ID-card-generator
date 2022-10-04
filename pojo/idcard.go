package pojo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IDCardGenerator struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name         string             `bson:"name,omitempty" json:"name,omitempty"`
	EmployeeId   string             `bson:"employeeId,omitempty" json:"employeeId,omitempty"`
	Age          string             `bson:"age,omitempty" json:"age,omitempty"`
	DateOfBirth  string             `bson:"date_of_birth,omitempty" json:"date_of_birth,omitempty"`
	Address      string             `bson:"address,omitempty" json:"address,omitempty"`
	Designation  string             `bson:"designation, omitempty" json:"designation,omitempty"`
	BloodGroup   string             `bson:"blood_group,omitempty" json:"blood_group,omitempty"`
	JoiningDate  string             `bson:"joining_date, omitempty" json:"joining_date,omitempty"`
	Date         time.Time          `bson:"date,omitempty" json:"date,omitempty"`
	FileLocation []string           `bson:"file_location,omitempty" json:"file_location,omitempty"`
}

type Search struct {
	Name        string `bson:"name,omitempty" json:"name,omitempty"`
	EmployeeId  string `bson:"employeeId,omitempty" json:"employeeId,omitempty"`
	JoiningDate string `bson:"joining_date, omitempty" json:"joining_date,omitempty"`
}
