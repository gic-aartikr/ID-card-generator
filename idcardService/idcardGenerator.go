package idcardService

import (
	"context"
	"errors"
	"fmt"
	"idCardDemo/pojo"
	"io"
	"log"
	"mime/multipart"
	"os"
	"strconv"

	"time"

	"github.com/unidoc/unipdf/v3/common/license"
	"github.com/unidoc/unipdf/v3/creator"
	"github.com/unidoc/unipdf/v3/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IDCardGenerator struct {
	Server     string
	Database   string
	Collection string
}

var Collection *mongo.Collection
var CollectionCategory *mongo.Collection
var ctx = context.TODO()
var insertDocs int

func (c *IDCardGenerator) Connect() {
	clientOptions := options.Client().ApplyURI(c.Server)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	Collection = client.Database(c.Database).Collection(c.Collection)

}

const maxUploadSize = 10 * 1024 * 1024 // 10 mb
const dir = "data/download/"

var fileName string

func (e *IDCardGenerator) InsertIDCardData(idCard pojo.IDCardGenerator, files []*multipart.FileHeader) (string, error) {
	var data []*pojo.IDCardGenerator
	var cursor *mongo.Cursor
	var err error
	// var idCardData pojo.IDCardGenerator
	arrFiles, err := uplaodFiles(files)

	if err != nil {
		return "", err
	}
	empname := idCard.Name
	dateOfBirth := idCard.DateOfBirth

	cursor, err = Collection.Find(ctx, bson.D{primitive.E{Key: "name", Value: empname}, primitive.E{Key: "date_of_birth", Value: dateOfBirth}})
	if err != nil {
		return "", errors.New("No data present in db for given empname and date of birth")
	}
	for cursor.Next(ctx) {
		var e pojo.IDCardGenerator
		err := cursor.Decode(&e)
		if err != nil {
			return "", err
		}
		data = append(data, &e)
	}

	if data == nil {
		addNewRecord(idCard, arrFiles)

		return " record add Successfully", nil
	} else {
		return "", errors.New("dublicate record")
	}

}

func addNewRecord(idCard pojo.IDCardGenerator, arrFiles []string) error {
	//  var cursor *mongo.Cursor
	var data *pojo.IDCardGenerator
	var empId string
	filter := []bson.M{{
		"$group": bson.M{
			"_id":        nil,
			"employeeId": bson.M{"$max": "$employeeId"},
		}},
	}
	cursor, err := Collection.Aggregate(ctx, filter)
	for cursor.Next(ctx) {
		var e pojo.IDCardGenerator
		err := cursor.Decode(&e)
		if err != nil {
			return err
		}
		data = &e
		empId = data.EmployeeId
		fmt.Println("empId:", empId)
	}

	fmt.Println("cursor:", cursor)
	idCard.Date = time.Now()
	num := 1
	intVar, err := strconv.Atoi(empId)
	emp := intVar + num
	empValue := strconv.Itoa(emp)
	idCard.EmployeeId = empValue
	idCard.FileLocation = arrFiles

	result, err := Collection.InsertOne(ctx, idCard)
	newID := result.InsertedID
	fmt.Println("newID:", newID)
	if err != nil {
		return errors.New("Unable To Insert New Record")
	}
	return err
}

func (e *IDCardGenerator) WriteIDCardDataInPDF(id string) error {
	var idCardData *pojo.IDCardGenerator
	// dir := "data/download/"
	file := "IDCardData" + fmt.Sprintf("%v", time.Now().Format("3_4_5_pm"))

	uniqueid, err := primitive.ObjectIDFromHex(id)
	fmt.Println("uniqueid:", uniqueid)
	if err != nil {
		return err
	}

	filter := bson.D{primitive.E{Key: "_id", Value: uniqueid}}

	resultData, err := Collection.Find(ctx, filter)
	if err != nil {
		return err
	}

	fmt.Println("data record:", resultData)
	for resultData.Next(ctx) {
		var data pojo.IDCardGenerator
		err := resultData.Decode(&data)
		if err != nil {
			return err
		}
		idCardData = &data

	}

	if idCardData == nil {
		return errors.New("Data Not Found In DB")
	}

	_, err = writeToPdf(dir, file, idCardData)

	if err != nil {
		return err
	}

	fmt.Println("idCardData:", idCardData)
	return nil
}

func writeToPdf(dir, file string, idCardData *pojo.IDCardGenerator) (*creator.Creator, error) {
	c := creator.New()
	err := license.SetMeteredKey("ac57153dc5ab976b46a5e8e8031c30a44153f513540f7a16db9e74293979494a")

	robotoFontRegular, err := model.NewPdfFontFromTTFFile("Roboto/Roboto-Regular.ttf")
	if err != nil {
		return c, err
	}

	// robotoFontPro, err := model.NewPdfFontFromTTFFile("Roboto/Roboto-Bold.ttf")
	if err != nil {
		return c, err
	}

	c.SetPageMargins(50, 50, 50, 50)

	normalFont := robotoFontRegular
	// normalFontColor := creator.ColorRGBFrom8bit(72, 86, 95)
	normalFontColorGreen := creator.ColorRGBFrom8bit(4, 79, 3)
	normalFontSize := 10.0

	iDTable := c.NewTable(2)
	issuerTable := c.NewTable(1)
	fmt.Println("FileLocation:", idCardData.FileLocation[0])

	//////////image///////////////////
	img, err := c.NewImageFromFile(idCardData.FileLocation[0])
	if err != nil {
		panic(err)
	}

	img.ScaleToHeight(200)
	img.ScaleToWidth(150)
	img.SetMargins(50, 50, 50, 50)

	p := c.NewParagraph("Name" + ":" + "  " + idCardData.Name)
	p.SetFont(normalFont)
	p.SetFontSize(normalFontSize)
	p.SetColor(normalFontColorGreen)
	p.SetMargins(0, 0, 5, 0)
	p.SetLineHeight(2)
	cell := issuerTable.NewCell()
	cell.SetBorder(creator.CellBorderSideAll, creator.CellBorderStyleSingle, 0)
	cell.SetContent(p)

	p = c.NewParagraph("Age" + ":" + "  " + idCardData.Age)
	p.SetFont(normalFont)
	p.SetFontSize(normalFontSize)
	p.SetColor(normalFontColorGreen)
	p.SetMargins(0, 0, 5, 0)
	p.SetLineHeight(2)
	cell = issuerTable.NewCell()
	cell.SetBorder(creator.CellBorderSideAll, creator.CellBorderStyleSingle, 0)
	cell.SetContent(p)

	p = c.NewParagraph("DateOfBirth" + ":" + "  " + idCardData.DateOfBirth)
	p.SetFont(normalFont)
	p.SetFontSize(normalFontSize)
	p.SetColor(normalFontColorGreen)
	p.SetMargins(0, 0, 5, 0)
	p.SetLineHeight(2)
	cell = issuerTable.NewCell()
	cell.SetBorder(creator.CellBorderSideAll, creator.CellBorderStyleSingle, 0)
	cell.SetContent(p)

	p = c.NewParagraph("EmployeeId" + ":" + "  " + idCardData.EmployeeId)
	p.SetFont(normalFont)
	p.SetFontSize(normalFontSize)
	p.SetColor(normalFontColorGreen)
	p.SetMargins(0, 0, 5, 0)
	p.SetLineHeight(2)
	cell = issuerTable.NewCell()
	cell.SetBorder(creator.CellBorderSideAll, creator.CellBorderStyleSingle, 0)
	cell.SetContent(p)

	p = c.NewParagraph("Address" + ":" + "  " + idCardData.Address)
	p.SetFont(normalFont)
	p.SetFontSize(normalFontSize)
	p.SetColor(normalFontColorGreen)
	p.SetMargins(0, 0, 5, 0)
	p.SetLineHeight(2)
	cell = issuerTable.NewCell()
	cell.SetBorder(creator.CellBorderSideAll, creator.CellBorderStyleSingle, 0)
	cell.SetContent(p)

	p = c.NewParagraph("Designation" + ":" + "  " + idCardData.Designation)
	p.SetFont(normalFont)
	p.SetFontSize(normalFontSize)
	p.SetColor(normalFontColorGreen)
	p.SetMargins(0, 0, 5, 0)
	p.SetLineHeight(2)
	cell = issuerTable.NewCell()
	cell.SetBorder(creator.CellBorderSideAll, creator.CellBorderStyleSingle, 0)
	cell.SetContent(p)

	p = c.NewParagraph("BloodGroup" + ":" + "  " + idCardData.BloodGroup)
	p.SetFont(normalFont)
	p.SetFontSize(normalFontSize)
	p.SetColor(normalFontColorGreen)
	p.SetMargins(0, 0, 5, 0)
	p.SetLineHeight(2)
	cell = issuerTable.NewCell()
	cell.SetBorder(creator.CellBorderSideAll, creator.CellBorderStyleSingle, 0)
	cell.SetContent(p)

	p = c.NewParagraph("JoiningDate" + ":" + "  " + idCardData.JoiningDate)
	p.SetFont(normalFont)
	p.SetFontSize(normalFontSize)
	p.SetColor(normalFontColorGreen)
	p.SetMargins(0, 0, 5, 0)
	p.SetLineHeight(2)

	cell = issuerTable.NewCell()
	cell.SetBorder(creator.CellBorderSideAll, creator.CellBorderStyleSingle, 0)
	cell.SetContent(p)

	idCell := iDTable.NewCell()
	idCell.SetBorder(creator.CellBorderSideAll, creator.CellBorderStyleSingle, 1)
	idCell.SetContent(issuerTable)

	idCardCell := iDTable.NewCell()
	idCardCell.SetBorder(creator.CellBorderSideAll, creator.CellBorderStyleSingle, 1)
	idCardCell.SetContent(img)
	c.Draw(iDTable)
	c.WriteToFile(dir + file + "report.pdf")
	fmt.Println("c:", c)
	return c, nil
}

func uplaodFiles(files []*multipart.FileHeader) ([]string, error) {

	var fileNames []string
	for _, fileHeader := range files {
		fileName = fileHeader.Filename
		fileNames = append(fileNames, dir+fileName)
		if fileHeader.Size > maxUploadSize {
			return fileNames, errors.New("The uploaded image is too big: %s. Please use an image less than 1MB in size: " + fileHeader.Filename)
		}

		file, err := fileHeader.Open()
		if err != nil {
			return fileNames, err
		}

		defer file.Close()

		buff := make([]byte, 512)
		_, err = file.Read(buff)
		if err != nil {
			return fileNames, err
		}

		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			return fileNames, err
		}

		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return fileNames, err
		}

		f, err := os.Create(dir + fileHeader.Filename)
		if err != nil {
			return fileNames, err
		}

		defer f.Close()

		_, err = io.Copy(f, file)
		if err != nil {
			return fileNames, err
		}
	}
	return fileNames, nil
}

func (e *IDCardGenerator) DeleteIDCard(id string) (string, error) {

	idcard, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return "", err
	}

	fmt.Println("idcard:", idcard)

	filter := bson.D{primitive.E{Key: "_id", Value: idcard}}

	fmt.Println("filter:", filter)

	cur, err := Collection.DeleteOne(ctx, filter)

	if err != nil {
		return "", err
	}

	if cur.DeletedCount == 0 {
		return "", errors.New("Unable To Delete Data")
	}

	return "Deleted Successfully", nil
}

func (e *IDCardGenerator) UpdateDataInIDcard(idcard pojo.IDCardGenerator, field string) (string, error) {

	id, err := primitive.ObjectIDFromHex(field)

	if err != nil {
		return "", err
	}

	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	fmt.Println("filter:", filter)
	update := bson.D{primitive.E{Key: "$set", Value: idcard}}

	err2 := Collection.FindOneAndUpdate(ctx, filter, update).Decode(e)

	if err2 != nil {
		return "", err2
	}
	return "Data Updated Successfully", nil
}

func (e *IDCardGenerator) SearchByNameEmployeeAndJoiningDate(cla pojo.Search) ([]*pojo.IDCardGenerator, error) {
	var recordData []*pojo.IDCardGenerator

	var cursor *mongo.Cursor
	var err error
	var str = ""
	var name = cla.Name
	var empId = cla.EmployeeId
	// var joiningDate = cla.JoiningDate

	if (name != "") && (empId != "") {

		cursor, err = Collection.Find(ctx, bson.D{primitive.E{Key: "name", Value: name}, {Key: "employeeId", Value: empId}})

		if err != nil {
			return recordData, err
		}

		str = "No data present in db for given name and employeeId "
	} else if (cla.Name) != "" {

		empName := cla.Name

		cursor, err = Collection.Find(ctx, bson.D{primitive.E{Key: "name", Value: empName}})

		if err != nil {
			return recordData, err
		}
		str = "No data present in db for given employee name"
	} else if (cla.EmployeeId) != "" {

		empId := cla.EmployeeId

		cursor, err = Collection.Find(ctx, bson.D{primitive.E{Key: "name", Value: empId}})

		if err != nil {
			return recordData, err
		}
		str = "No data present in db for given employee id"
	} else {
		fmt.Println("city:", cla.JoiningDate)
		cursor, err = Collection.Find(ctx, bson.D{primitive.E{Key: "joining_date", Value: cla.JoiningDate}})

		if err != nil {
			return recordData, err
		}
		str = "No  data present in db for given joining date"
	}

	for cursor.Next(ctx) {
		var e pojo.IDCardGenerator
		err := cursor.Decode(&e)
		if err != nil {
			return recordData, err
		}
		recordData = append(recordData, &e)
	}

	if recordData == nil {
		return recordData, errors.New(str)
	}

	if err != nil {
		fmt.Println(err)
	}
	return recordData, err
}
