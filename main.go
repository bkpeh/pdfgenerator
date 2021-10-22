package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-pdf/fpdf"
)

var filename [5]string = [5]string{"WGS Secondary 1 Books Purchase.csv", "WGS Secondary 2 Books Purchase.csv",
	"WGS Secondary 3 Books Purchase.csv", "WGS Secondary 4 Books Purchase.csv", "WGS Secondary 5 Books Purchase.csv"}

var savepath [5]string = [5]string{"Sec1", "Sec2", "Sec3", "Sec4", "Sec5"}

//Total of 42 columns for Sec 2
//Total of 54 columns for Sec 3
//Total of 47 columns for Sec 4
var lastdatacol [5]int = [5]int{38, 38, 50, 43, 38} //Ignore the last 3 column which is name, mobile and address

const csvdir string = "csv"
const pdfdir string = "pdf"
const delivery float32 = 10.0

//Argument 1-5 and 99 is valid input. 99 is to process Sec 1 to 5
func main() {

	input, err := strconv.Atoi(os.Args[1])

	if err != nil {
		log.Fatalln("Argument is not a number.  ", err)
	} else {

		fmt.Println("Argument : ", os.Args[1])

		if input == 99 {
			for level := 0; level < 5; level++ {
				openCSV(level)
			}
		} else if input > 0 && input < 6 {
			openCSV(input - 1)
		} else {
			fmt.Println("Argument out of range.")
		}
	}
}

func openCSV(level int) {

	var invoicenum string
	var rownum int
	var imgopt fpdf.ImageOptions

	file, err := os.Open(filepath.Join(csvdir, filename[level]))

	fmt.Println("filename : ", filepath.Join(csvdir, filename[level]))

	if err != nil {
		fmt.Println("Cannot open CSV file. ", err)
		return
	}

	reader := csv.NewReader(file)

	if reader == nil {
		fmt.Println("Cannot read CSV file. ")
		return
	}

	//First 2 column not needed
	//Total of 42 columns for Sec 2
	//Total of 54 columns for Sec 3
	//Total of 47 columns for Sec 4

	reader.Read() //To read the column header record and ignore

	rownum = 2 //Count the row. Start at 2 cause 1 is column header

	for {
		record, err := reader.Read()

		if err != nil {
			if err == io.EOF {
				fmt.Println("EOF")
				break
			} else {
				fmt.Println("Error reading record. ", err)
			}
		}

		colsize := len(record)
		name := record[colsize-4]
		contact := "(" + record[colsize-3] + ")"
		postal := getPostalCode(record[colsize-2])
		address := record[colsize-1]
		email := record[1]

		fmt.Printf("Name: %s, Postal Code: %s, Address: %s, Size: %d\n", name, postal, address, colsize)

		//Config the PDF file
		pdf := fpdf.New("P", "mm", "A4", "")
		pdf.SetFont("Arial", "B", 22)
		pdf.AddPage()

		//Company name and Logo
		imgopt.ImageType = "jpg"
		pdf.ImageOptions("wg_logo.jpg", pdf.GetY()+160, 5, 30, 30, false, imgopt, 0, "")
		pdf.CellFormat(10, 10, "Woodgrove Stationery and Bookshop", "", 1, "L", false, 0, "")
		pdf.Ln(5)
		pdf.SetFont("Arial", "B", 12)

		//To: section
		pdf.CellFormat(5, 5, "To:", "", 1, "L", false, 0, "")
		pdf.SetFont("Arial", "", 12)

		//Print Name
		pdf.CellFormat(5, 5, strings.ToUpper(name), "", 1, "L", false, 0, "")

		//Print Address
		pdf.CellFormat(5, 5, strings.Title(address), "", 1, "L", false, 0, "")

		//Print Postal Code
		pdf.CellFormat(5, 5, postal, "", 1, "L", false, 0, "")

		//Print Email
		pdf.CellFormat(5, 5, email, "", 1, "L", false, 0, "")

		//Print Contact
		pdf.CellFormat(5, 5, contact, "", 1, "L", false, 0, "")

		var total float32

		for i, v := range record {
			if i == 0 {
				date := v[:strings.Index(v, " ")]
				date = strings.ReplaceAll(date, `/`, "")

				invoicenum = strconv.Itoa(rownum) + "_" + strconv.Itoa(level+1) + "_" + date

				//Print Invoice Number
				pdf.Ln(5)
				pdf.SetFont("Arial", "B", 12)
				pdf.CellFormat(5, 5, "Invoice No. : "+invoicenum, "", 1, "L", false, 0, "")
				pdf.CellFormat(5, 5, "PayNow To : 86111871 (Cheng Bee Lian)", "", 1, "L", false, 0, "")
			}

			if i == 2 {
				pdf.SetFont("Arial", "", 12)
				pdf.Ln(5)
				pdf.CellFormat(5, 5, "Stream: "+v, "", 1, "L", false, 0, "")
				pdf.Line(pdf.GetX(), pdf.GetY(), pdf.GetX()+190, pdf.GetY())
				pdf.CellFormat(5, 5, "Description"+genSpace(110)+"Price", "", 1, "L", false, 0, "") //Print table column header
				pdf.Line(pdf.GetX(), pdf.GetY(), pdf.GetX()+190, pdf.GetY())
				pdf.Ln(3)
			}

			pdf.SetFont("Arial", "", 12)

			//Multi data in 1 record. Etc Maths Textbook 1A;Maths Workbook 1A
			elements := strings.Split(v, ";")
			fmt.Printf("Len : %d, Data: %v\n", len(elements), elements)

			if len(elements) != 0 {
				for _, vv := range elements {
					if getPrice(vv) == -1 {
						fmt.Println("getPrice :-1")
						continue
					}

					//Add the items into pdf
					if i > 2 && i < lastdatacol[level] && vv != "" {
						total = total + addCell(pdf, vv)
					}
				}
			} else { //Record only has single data
				if getPrice(v) == -1 {
					fmt.Println("getPrice :-1")
					continue
				}

				//Add the items into pdf
				if i > 2 && i < lastdatacol[level] && v != "" {
					total = total + addCell(pdf, v)
				}
			}
		}

		//Add Delivery charge
		pdf.Ln(5)
		pdf.CellFormat(5, 5, genSpace(104)+"Delivery Cost : 10.00", "", 1, "L", false, 0, "")

		//Add total to pdf
		totalstr := fmt.Sprintf("%.2f", total+delivery)
		pdf.Ln(1)
		pdf.CellFormat(5, 5, genSpace(117)+"Total : $"+totalstr, "", 1, "L", false, 0, "")

		//Create the directory if it doesn't exist yet
		path := filepath.Join(pdfdir, savepath[level])
		err = os.MkdirAll(path, os.ModePerm)

		if err != nil {
			fmt.Println("Error creating directory. ", err)
		}

		//Full path with filename
		path = filepath.Join(path, invoicenum+"_"+name+".pdf")

		//fmt.Println("Save path : ", path)

		//Create and save pdf. PDF filename is the name in csv
		err = pdf.OutputFileAndClose(path)

		if err != nil {
			fmt.Println("Error creating PDF. ", err)
		}

		rownum++
	}
}

//Get the price from the description text
func getPrice(text string) float32 {

	var startChar string = "$"
	var endChar string = ")"

	first := strings.Index(text, startChar)
	second := strings.LastIndex(text, endChar)

	if first == -1 || second == -1 {
		fmt.Println("Character $ or ) not found")
	} else if second > first {
		//fmt.Println("Trim : ", strings.TrimRight(text, startChar))
		fmt.Println("Price : ", text[first+1:second])

		price, err := strconv.ParseFloat(text[first+1:second], 32)

		if err != nil {
			fmt.Println("Not a valid price")
		} else {
			return float32(price)
		}
	}

	return -1.0
}

//Generate space for padding
func genSpace(space int) string {
	var pad string

	for i := 0; i < space; i++ {
		pad = pad + " "
	}

	return pad
}

func addCell(pdf *fpdf.Fpdf, record string) float32 {
	//fmt.Println("Index", i, "Record : ", vv)
	//total = total + getPrice(vv)

	pricestr := fmt.Sprintf("%.2f", getPrice(record))

	fmt.Println("Len : ", len(record))

	record = record[:strings.LastIndex(record, "(")]

	//If description is too long
	if len(record) > 75 {
		index := strings.Index(record, ":")

		if index == -1 {
			index = strings.Index(record, ",")
		}

		//Breakdown description into 2 line. ":" is the token
		pdf.CellFormat(5, 5, record[:index+1], "", 0, "L", false, 0, "")       //Add description before ":"
		pdf.CellFormat(5, 5, genSpace(124)+pricestr, "", 1, "L", false, 0, "") //Add price
		//pdf.Ln(1)
		pdf.CellFormat(5, 5, strings.Trim(record[index+1:], " "), "", 1, "L", false, 0, "") //Add description after ":" on next line
	} else {
		pdf.CellFormat(5, 5, record, "", 0, "L", false, 0, "")                 //Add description
		pdf.CellFormat(5, 5, genSpace(124)+pricestr, "", 1, "L", false, 0, "") //Add price
	}

	fvalue, err := strconv.ParseFloat(pricestr, 32)

	if err != nil {
		fmt.Println("Cannot get float value")
		fvalue = -1
	}

	return float32(fvalue)
}

func getPostalCode(code string) string {
	exp := regexp.MustCompile(`\d`)

	postalcode := exp.FindAllString(code, -1)

	return strings.Join(postalcode, "")
}

/*
func getDeliveryPrice(code string) float32 {
	postalcode := getPostalCode(code)

	if postalcode[:2] == "72" || postalcode[:2] == "73" {
		return 10.0
	}

	return 15.0
}
*/
