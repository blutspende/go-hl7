package hl7

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"
)

// Marshal - wrap datastructure to code
func Marshal(message interface{}, fieldSeparator FieldSeparator, enc Encoding, tz Timezone, notation Notation) ([][]byte, error) {
	location, err := time.LoadLocation(string(tz))
	if err != nil {
		return [][]byte{}, err
	}

	var delimiters Delimiters
	delimiters.Composite = string(fieldSeparator)
	delimiters.Sub = "^"
	delimiters.Repeat = "~"
	delimiters.Escape = "\\"
	delimiters.SubSub = "&"

	buffer, err := processStruct(message, 1, enc, location, notation, delimiters)

	return buffer, err
}

type OutputRecord struct {
	Field, Repeat, Component int
	Value                    string
}

type OutputRecords []OutputRecord

func processStruct(message interface{}, depth int, enc Encoding, location *time.Location, notation Notation, delimiters Delimiters) ([][]byte, error) {
	buffer := make([][]byte, 0)
	messageValue := reflect.ValueOf(message)
	messageType := reflect.TypeOf(message)

	for i := 0; i < messageValue.NumField(); i++ {
		currentRecord := messageValue.Field(i)
		currentRecordTag := messageType.Field(i).Tag.Get("hl7")
		currentRecordTagsList := strings.Split(currentRecordTag, ",")

		if len(currentRecordTag) == 0 { // no annotation = Descend if its an array or a struct of such
			if currentRecord.Kind() == reflect.Slice { // array of something = iterate and recurse
				for x := 0; x < currentRecord.Len(); x++ {
					dood := currentRecord.Index(x).Interface()
					if bytes, err := processStruct(dood, depth+1, enc, location, notation, delimiters); err != nil {
						return nil, err
					} else {
						for line := 0; line < len(bytes); line++ {
							buffer = append(buffer, bytes[line])
						}
					}
				}

			} else if currentRecord.Kind() == reflect.Struct { // got the struct straignt = recurse directly
				if bytes, err := processStruct(currentRecord.Interface(), depth+1, enc, location, notation, delimiters); err != nil {
					return nil, err
				} else {
					for line := 0; line < len(bytes); line++ {
						buffer = append(buffer, bytes[line])
					}
				}
			} else {
				return nil, fmt.Errorf("invalid Datatype without any annotation '%s'. You can use struct or slices of structs", currentRecord.Kind())
			}
		} else { // an annotated Record
			recordType := currentRecordTagsList[0]

			switch currentRecord.Kind() {
			case reflect.Slice: // it is an annotated slice
				if !currentRecord.IsNil() {
					for x := 0; x < currentRecord.Len(); x++ {
						outs, err := processSegment(recordType, 0, currentRecord.Index(x), x+1, location, delimiters) // fmt.Println(outp)
						if err != nil {
							return nil, err
						}
						buffer = append(buffer, []byte(outs))
					}
				}
			case reflect.Struct:
				outs, err := processSegment(recordType, 0, currentRecord, 1, location, delimiters) // fmt.Println(outp)
				if err != nil {
					return nil, err
				}
				if outs != "" {
					buffer = append(buffer, []byte(outs))
				}
			default:
				return buffer, fmt.Errorf("invalid Datatype of nested structure type:%s annotation:%s - abort",
					currentRecord.Type().Name(), currentRecordTag)
			}
		}
	}

	return buffer, nil
}

func processSegment(recordType string, subDepth int, currentRecord reflect.Value, generatedSequenceNumber int, location *time.Location, delimiters Delimiters) (string, error) {
	fieldList := make(OutputRecords, 0)
	isWorthGeneratingThisRecord := false

	for i := 0; i < currentRecord.NumField(); i++ {
		field := currentRecord.Field(i)
		fieldTag := currentRecord.Type().Field(i).Tag.Get("hl7")
		fieldTagsList := strings.Split(fieldTag, ",")

		fieldIdx, repeatIdx, componentIdx, err := readFieldAddressAnnotation(fieldTagsList[0])
		if err != nil {
			return "", fmt.Errorf("invalid annotation for field %s : (%s)", field.Type().Name(), err)
		}

		// fmt.Printf("Decode %+v to %d.%d.%d for %s\n", fieldAstmTagsList, fieldIdx, repeatIdx, componentIdx, field.String())

		switch field.Type().Kind() {
		case reflect.String:
			value := ""

			if sliceContainsString(fieldTagsList, ANNOTATION_SEQUENCE) {
				return "", fmt.Errorf("invalid annotation %s for string-field", ANNOTATION_SEQUENCE)
			}

			// if no delimiters are given, default is \^&
			if sliceContainsString(fieldTagsList, ANNOTATION_DELIMITER) && field.String() == "" {
				value = delimiters.Sub + delimiters.Repeat + delimiters.Escape + delimiters.SubSub
			} else {
				value = field.String()
			}

			if value != "" {
				isWorthGeneratingThisRecord = true
			}

			fieldList = addFieldToOutput(fieldList, fieldIdx, repeatIdx, componentIdx, value)
		case reflect.Int:
			value := fmt.Sprintf("%d", field.Int())
			if sliceContainsString(fieldTagsList, ANNOTATION_SEQUENCE) {
				value = fmt.Sprintf("%d", generatedSequenceNumber)
				generatedSequenceNumber = generatedSequenceNumber + 1
			}
			fieldList = addFieldToOutput(fieldList, fieldIdx, repeatIdx, componentIdx, value)
		case reflect.Float32, reflect.Float64:
			//TODO: Add annotation for amount of decimals
			value := fmt.Sprintf("%.3f", field.Float())
			fieldList = addFieldToOutput(fieldList, fieldIdx, repeatIdx, componentIdx, value)
		case reflect.Slice:
			value := ""
			for i := 0; i < field.Len(); i++ {
				var oneElementStr string
				var err error
				switch field.Index(i).Kind() {
				case reflect.Int:
				case reflect.String:
					oneElementStr = field.Index(i).String()
				case reflect.Struct:
					oneElementStr, err = processSegment("", subDepth+1, field.Index(i), 0, location, delimiters)
				default:
					oneElementStr = "not implemented"
				}
				if err != nil {
					return "", err
				}
				if i > 0 {
					value = value + "~"
				}
				value = value + oneElementStr
			}
			if value != "" {
				isWorthGeneratingThisRecord = true
			}
			fieldList = addFieldToOutput(fieldList, fieldIdx, repeatIdx, componentIdx, value)
		case reflect.Struct:
			if field.Type().Name() == "Time" { // ToDo: Ambigious Time (time.Time or sthelse.Time ?)
				time := field.Interface().(time.Time)
				if !time.IsZero() {
					isWorthGeneratingThisRecord = true
					if sliceContainsString(fieldTagsList, ANNOTATION_LONGDATE) {
						value := time.In(location).Format("20060102150405")
						fieldList = addFieldToOutput(fieldList, fieldIdx, repeatIdx, componentIdx, value)
					} else { // short date
						value := time.In(location).Format("20060102")
						fieldList = addFieldToOutput(fieldList, fieldIdx, repeatIdx, componentIdx, value)
					}
				}
			} else { // structure
				//TODO:Limit depth to go
				subbie, _ := processSegment("", subDepth+1, field, 0, location, delimiters)
				if subbie != "" {
					fieldList = addFieldToOutput(fieldList, fieldIdx, repeatIdx, componentIdx, subbie)
					isWorthGeneratingThisRecord = true
				}
			}

		default:
			return "", fmt.Errorf("invalid field type '%s' in struct '%s', input not processed", field.Type().Name(), currentRecord.Type().Name())
		}
	}

	generatePreceedingAndTrailingDelimiters := true
	thisdelimiter := "?"
	nextdelimiter := "?"
	switch subDepth {
	case 0:
		generatePreceedingAndTrailingDelimiters = true
		thisdelimiter = delimiters.Composite
		nextdelimiter = delimiters.Sub
	case 1:
		generatePreceedingAndTrailingDelimiters = false
		thisdelimiter = delimiters.Sub
		nextdelimiter = delimiters.SubSub
	case 2:
		generatePreceedingAndTrailingDelimiters = false
		thisdelimiter = delimiters.SubSub
		thisdelimiter = ""
	}
	str, hasData := generateHL7String(recordType, fieldList, delimiters,
		generatePreceedingAndTrailingDelimiters, thisdelimiter, nextdelimiter)
	// fmt.Printf("Build string '%s' (%d)\n", str, subDepth)

	// if there is no data, or the date presented is empty were not generating this record
	if !hasData || !isWorthGeneratingThisRecord {
		return "", nil
	}

	// Fields with ^ or & are always right-trimmed
	if subDepth > 0 {
		str = strings.TrimRight(str, thisdelimiter)
	}

	return str, nil
}

func addFieldToOutput(data []OutputRecord, field, repeat, component int, value string) []OutputRecord {
	or := OutputRecord{
		Field:     field,
		Component: component,
		Value:     value,
	}

	data = append(data, or)
	return data
}

// used for sorting
func (or OutputRecords) Len() int { return len(or) }
func (or OutputRecords) Less(i, j int) bool {
	if or[i].Field == or[j].Field {
		if or[i].Repeat == or[j].Repeat {
			return or[i].Component < or[j].Component
		} else {
			return or[i].Repeat < or[j].Repeat
		}
	} else {
		return or[i].Field < or[j].Field
	}
}

func (or OutputRecords) Swap(i, j int) { or[i], or[j] = or[j], or[i] }

func generateHL7String(recordtype string, fieldList OutputRecords, delimiters Delimiters, generatePreceedingAndTrailingDelimiters bool, THISDELIMTER, NEXTDELIMITER string) (string, bool) {
	// sort by Field index, keeping original order or equal elements
	// we need to fill empty sections later; see inner section of the main loop
	sort.SliceStable(fieldList, func(i, j int) bool {
		return fieldList[i].Field < fieldList[j].Field
	})

	// if we have a record type e.g. MSH then append it
	output := ""
	if generatePreceedingAndTrailingDelimiters {
		output += recordtype + THISDELIMTER
	}

	// main loop to build the output string
	lastFieldId := 0
	for i, field := range fieldList {
		// do not append FieldSeparator to the output (it's a special field in the HL7)
		if recordtype == "MSH" && i == 0 {
			continue
		}

		// sometimes there is an empty space in the fieldList e.g. last one was 4 and the next one is 6
		if lastFieldId+1 != field.Field {
			// fill these fields with empty sections
			for j := lastFieldId + 1; j < field.Field; j++ {
				output += THISDELIMTER
			}
		}
		lastFieldId = field.Field

		// append data
		output += field.Value

		// append delimiter
		if i < len(fieldList)-1 {
			if fieldList[i+1].Component != 0 {
				output += NEXTDELIMITER
			} else {
				output += THISDELIMTER
			}
		}
	}

	// remove empty fields from the end
	lastFieldIdxWithData := 0
	outputParts := strings.Split(output, delimiters.Composite)
	for b := len(outputParts) - 1; b >= 0; b-- {
		if outputParts[b] != "" && outputParts[b] != "0" {
			lastFieldIdxWithData = b
			break
		}
	}
	output = ""
	for i := 0; i <= lastFieldIdxWithData; i++ {
		output += outputParts[i]

		if i < lastFieldIdxWithData {
			output += delimiters.Composite
		}
	}

	// if its end of the segment then append a CR
	if generatePreceedingAndTrailingDelimiters {
		if recordtype == "MSH" {
			output += delimiters.Composite
		}
		output += "\r"
	}

	return output, true
}
