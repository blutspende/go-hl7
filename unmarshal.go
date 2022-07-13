package hl7

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
)

// HL7 Format delimiters
// https://blog.interfaceware.com/hl7-delimiter-redefinitions/
type Delimiters struct {
	Composite string
	Sub       string
	SubSub    string
	Repeat    string
	Escape    string
}

type Error string

var EOF = fmt.Errorf("EOF")

func IdentifyMessage(messageData []byte, encoding Encoding) (messageType string, protocolVersion string, err error) {
	var messageBytes []byte

	if messageBytes, err = encodeMessage(messageData, encoding); err != nil {
		return "", "", err
	}
	messageString := string(messageBytes)

	if len(messageString) < 4 {
		return "", "", fmt.Errorf("invalid messageData='%s'", messageString)
	}

	if !strings.HasPrefix(messageString, "MSH") {
		return "", "", fmt.Errorf("expected segment was MSH current segment is '%s'", messageString[0:3])
	}
	separator := string(messageString[3])

	minSegmentPartCount := 12
	messageStringParts := strings.Split(messageString, separator)
	if len(messageStringParts) < minSegmentPartCount {
		return "", "", fmt.Errorf("expected segment length was %d, current length is '%d'", minSegmentPartCount, len(messageStringParts))
	}

	messageType = string(messageStringParts[8])
	protocolVersion = string(messageStringParts[11])

	return messageType, protocolVersion, err
}

func encodeMessage(messageData []byte, enc Encoding) ([]byte, error) {
	var messageBytes []byte
	var err error

	switch enc {
	case EncodingUTF8:
		return messageData, nil
	case EncodingASCII:
		return messageData, nil
	case EncodingDOS866:
		messageBytes, err = EncodeCharsetToUTF8From(charmap.CodePage866, messageData)
		return messageBytes, err
	case EncodingDOS855:
		messageBytes, err = EncodeCharsetToUTF8From(charmap.CodePage855, messageData)
		return messageBytes, err
	case EncodingDOS852:
		messageBytes, err = EncodeCharsetToUTF8From(charmap.CodePage852, messageData)
		return messageBytes, err
	case EncodingWindows1250:
		messageBytes, err = EncodeCharsetToUTF8From(charmap.Windows1250, messageData)
		return messageBytes, err
	case EncodingWindows1251:
		messageBytes, err = EncodeCharsetToUTF8From(charmap.Windows1251, messageData)
		return messageBytes, err
	case EncodingWindows1252:
		messageBytes, err = EncodeCharsetToUTF8From(charmap.Windows1252, messageData)
		return messageBytes, err
	case EncodingISO8859_1:
		messageBytes, err = EncodeCharsetToUTF8From(charmap.ISO8859_1, messageData)
		return messageBytes, err
	}

	return []byte{}, fmt.Errorf("invalid Codepage Id='%d' - %w", enc, err)
}

func Unmarshal(messageData []byte, targetStruct interface{}, enc Encoding, tz Timezone) error {
	var (
		messageBytes []byte
		err          error
	)

	if messageBytes, err = encodeMessage(messageData, enc); err != nil {
		return err
	}

	// first try to break by 0x0a (non-standard, but used sometimes)
	bufferedInputLinesWithBlanks := strings.Split(string(messageBytes), string([]byte{0x0A}))
	if len(bufferedInputLinesWithBlanks) <= 1 { // if it was not possible to break with non-standard 0x0a line-break try 0d (standard)
		bufferedInputLinesWithBlanks = strings.Split(string(messageBytes), string([]byte{0x0D}))
	}
	bufferedInputLines := make([]string, 0)
	for _, x := range bufferedInputLinesWithBlanks {
		trimmed := strings.Trim(x, " ")
		if trimmed != "" {
			bufferedInputLines = append(bufferedInputLines, trimmed)
		}
	}

	// strip the remaining 0A and 0D Linefeed at the end
	for i := 0; i < len(bufferedInputLines); i++ {
		// 0d,0a then again as there have been files observed which had 0a0d (0d0a would be normal)
		bufferedInputLines[i] = strings.Trim(bufferedInputLines[i], string([]byte{0x0A}))
		bufferedInputLines[i] = strings.Trim(bufferedInputLines[i], string([]byte{0x0D}))
		bufferedInputLines[i] = strings.Trim(bufferedInputLines[i], string([]byte{0x0A}))
		bufferedInputLines[i] = strings.Trim(bufferedInputLines[i], string([]byte{0x0D}))
		// fmt.Println(">", bufferedInputLines[i])
	}

	var delimiters Delimiters

	delimiters.Composite = "|"
	delimiters.Sub = "^"
	delimiters.Repeat = "~"
	delimiters.Escape = "\\"
	delimiters.SubSub = "&"

	if bufferedInputLines[0][0:3] == "MSH" {
		delimiters.Composite = bufferedInputLines[0][3:4] // Default |
		delimiters.Sub = bufferedInputLines[0][4:5]       // Default ^
		delimiters.Repeat = bufferedInputLines[0][5:6]    // Default ~
		delimiters.Escape = bufferedInputLines[0][6:7]    // Default \
		delimiters.SubSub = bufferedInputLines[0][7:8]    // Default &
	}

	var lastProcessedLine int
	if lastProcessedLine, _, err = ParseStruct(bufferedInputLines, 0 /*recursion*/, 0 /*line*/, targetStruct, enc, tz, delimiters); err != nil {
		if err != EOF {
			return err
		}
	}
	if lastProcessedLine != len(bufferedInputLines) {
		return fmt.Errorf("unable to process record %s", bufferedInputLines[lastProcessedLine])
	}

	return nil
}

func EncodeCharsetToUTF8From(charmap *charmap.Charmap, data []byte) ([]byte, error) {
	sr := bytes.NewReader(data)
	e := charmap.NewDecoder().Reader(sr)
	bytes := make([]byte, len(data)*2)
	n, err := e.Read(bytes)
	if err != nil {
		return []byte{}, err
	}
	return bytes[:n], nil
}

type RETV int

const (
	OK         RETV = 1
	UNEXPECTED RETV = 2 // an exit that wont abort processing. used for skipping optional records
	ERROR      RETV = 3 // a definite error that stops the process
)

// HL7 delimimters (default, rewritten at startup)
// https://blog.interfaceware.com/hl7-delimiter-redefinitions/
// This function takes a string and a struct and matches the annotated fields to the string-input
func ParseStruct(bufferedInputLines []string, depth int, currentInputLine int, targetStruct interface{}, enc Encoding, tz Timezone, delimiters Delimiters) (int, RETV, error) {

	hasMatchedAtLeastOneRecord := false

	timeLocation, err := time.LoadLocation(string(tz))
	if err != nil {
		return currentInputLine, ERROR, err
	}

	targetStructType := reflect.TypeOf(targetStruct).Elem()
	targetStructValue := reflect.ValueOf(targetStruct).Elem()

	for i := 0; i < targetStructType.NumField(); i++ {

		if currentInputLine >= len(bufferedInputLines) {
			break
		}

		currentRecord := targetStructValue.Field(i)
		ftype := targetStructType.Field(i)
		hl7 := ftype.Tag.Get("hl7")
		tagsList := removeEmptyStrings(strings.Split(hl7, ","))

		if len(tagsList) < 1 { // if it is not annotated at all its a "struct or a slice of structs to group records"

			switch currentRecord.Kind() {
			case reflect.Struct:

				var err error
				var retv RETV

				lineBefore := currentInputLine
				currentInputLine, retv, err = ParseStruct(bufferedInputLines, depth+1, currentInputLine, currentRecord.Addr().Interface(), enc, tz, delimiters)
				if currentInputLine != lineBefore {
					hasMatchedAtLeastOneRecord = true
				}

				if err != nil {
					if retv == UNEXPECTED {
						if depth > 0 {
							// if nested structures abort due to unexpected records that does not create an error
							// as the parse will be continued one level higher
							break
						} else {
							//It is wrong to exit here, there are maybe more fields that might match
							// return currentInputLine, ERROR, err
							// continue
						}
					}
					if retv == ERROR { // a serious error ends the processing
						return currentInputLine, ERROR, err
					}
				}

			case reflect.Slice:
				// Not annotated array. If it's a struct have to recurse, otherwise skip
				if targetStructType.Field(i).Type.Kind() == reflect.Slice {
					// Array of Structs
					if reflect.TypeOf(targetStructValue.Interface()).Kind() == reflect.Struct {
						innerStructureType := targetStructType.Field(i).Type.Elem()

						sliceForNestedStructure := reflect.MakeSlice(targetStructType.Field(i).Type, 0, 0)

						for {
							allocatedElement := reflect.New(innerStructureType)
							var err error
							var retv RETV
							lineBefore := currentInputLine
							currentInputLine, retv, err = ParseStruct(bufferedInputLines, depth+1,
								currentInputLine, allocatedElement.Interface(), enc, tz, delimiters)
							if lineBefore != currentInputLine {
								hasMatchedAtLeastOneRecord = true
							}
							if err != nil {
								if retv == UNEXPECTED {
									if depth > 0 {
										// if nested structures abort due to unexpected records that does not create an error
										// as the parse will be continued one level higher
										break
									} else {
										return currentInputLine, ERROR, err
									}
								}
								if retv == ERROR { // a serious error ends the processing
									return currentInputLine, ERROR, err
								}
							}

							sliceForNestedStructure = reflect.Append(sliceForNestedStructure, allocatedElement.Elem())
							reflect.ValueOf(targetStruct).Elem().Field(i).Set(sliceForNestedStructure)
						}
					}
				}
			}

			continue
		}

		expectInputRecordType := tagsList[0][0:3] // Expected Record type
		expectedInputRecordTypeOptional := false
		if sliceContainsString(tagsList, ANNOTATION_OPTIONAL) {
			expectedInputRecordTypeOptional = true
		}

		if currentInputLine >= len(bufferedInputLines) {
			break
		}

		if len(bufferedInputLines[currentInputLine]) == 0 {
			continue // skip empty line
		}

		if expectInputRecordType == bufferedInputLines[currentInputLine][0:3] {

			hasMatchedAtLeastOneRecord = true

			// Array of Segments
			if currentRecord.Kind() == reflect.Slice {

				innerStructureType := targetStructType.Field(i).Type.Elem()
				sliceForNestedStructure := reflect.MakeSlice(targetStructType.Field(i).Type, 0, 0)
				for { // iterate for as long as the same type repeats

					allocatedElement := reflect.New(innerStructureType)

					if err = parseStructA(bufferedInputLines[currentInputLine], 0 /*subdepth=0*/, allocatedElement.Elem(), timeLocation, delimiters); err != nil {
						return currentInputLine, ERROR, fmt.Errorf("failed to process input line '%s' err:%w", bufferedInputLines[currentInputLine], err)
					}

					sliceForNestedStructure = reflect.Append(sliceForNestedStructure, allocatedElement.Elem())
					reflect.ValueOf(targetStruct).Elem().Field(i).Set(sliceForNestedStructure)

					// keep reading while same elements are up
					currentInputLine = currentInputLine + 1
					if currentInputLine < len(bufferedInputLines) && expectInputRecordType != bufferedInputLines[currentInputLine][0:3] {
						break
					}

					if currentInputLine >= len(bufferedInputLines) {
						break
					}

				}

			} else {

				if err = parseStructA(bufferedInputLines[currentInputLine], 0 /*subdepth=0*/, currentRecord, timeLocation, delimiters); err != nil {
					return currentInputLine, ERROR, fmt.Errorf("failed to process input line '%s' err:%w", bufferedInputLines[currentInputLine], err)
				}

				currentInputLine = currentInputLine + 1

			}

		} else { // The expected input-record did not occur

			if expectedInputRecordTypeOptional {
				continue // skipping optional record instead of an error
			} else {
				return currentInputLine, UNEXPECTED, fmt.Errorf("expected Record-Type '%s' input was '%s' in depth (%d) (Abort)", expectInputRecordType, bufferedInputLines[currentInputLine][0:3], depth)
			}

		}

		if currentInputLine >= len(bufferedInputLines) { // EOF ?
			break
		}
	}

	if !hasMatchedAtLeastOneRecord {
		if currentInputLine < len(bufferedInputLines) {
			return currentInputLine, UNEXPECTED, fmt.Errorf("unable to process line '%s'", bufferedInputLines[currentInputLine])
		} else {
			return currentInputLine, UNEXPECTED, EOF
		}
	}
	return currentInputLine, OK, nil
}

func parseStructA(inputStr string, subdepth int, record reflect.Value, timezone *time.Location, delimiters Delimiters) error {

	delimiterBySubDepth := ""
	nextDelimiterBySubDepth := ""
	switch subdepth {
	case 0:
		delimiterBySubDepth = delimiters.Composite
		nextDelimiterBySubDepth = delimiters.Sub
	case 1:
		delimiterBySubDepth = delimiters.Sub
		nextDelimiterBySubDepth = delimiters.SubSub
	case 2:
		delimiterBySubDepth = delimiters.SubSub
	}

	// fmt.Printf("  Segment(%s: %d,%s) '%s'\n", record.Type().Name(), subdepth, delimiterBySubDepth, inputStr)

	if reflect.ValueOf(record).Type().Kind() != reflect.Struct {
		return fmt.Errorf("invalid type of target: '%s', expecting 'struct'", reflect.ValueOf(record).Type().Kind())
	}

	inputFields := strings.Split(inputStr, delimiterBySubDepth)
	if len(inputFields) < 1 {
		return fmt.Errorf("input contains no data or can not be split %s", inputStr)
	}

	for j := 0; j < record.NumField(); j++ {

		recordfield := record.Field(j)

		hl7Tag := record.Type().Field(j).Tag.Get("hl7")
		if hl7Tag == "" {
			continue // nothing to process when someone requires astm:
		}

		if !recordfield.CanInterface() {
			return fmt.Errorf("field %s is not exported - aborting import", recordfield.Type().Name())
		}
		recordFieldInterface := recordfield.Addr().Interface()

		hasOverrideDelimiterAnnotation := false

		hl7TagsList := strings.Split(hl7Tag, ",")
		for i := 0; i < len(hl7TagsList); i++ {
			hl7TagsList[i] = strings.Trim(hl7TagsList[i], " ")
		}
		if sliceContainsString(hl7TagsList, ANNOTATION_DELIMITER) {
			// the delimiter is instantly replaced with the delimiters from the file for further parsing. By default that is "\^&"
			hasOverrideDelimiterAnnotation = true
		}

		inputIsRequired := false

		addr_component, addr_repeat, addr_subcomponent, err := readFieldAddressAnnotation(hl7TagsList[0])
		if err != nil {
			return fmt.Errorf("invalid annotation for field %s. (%w)", record.Type().Field(j).Name, err)
		}
		if addr_component >= len(inputFields) || addr_component < 0 {
			continue // mapped field is beyond the data (not an error)
		}

		switch reflect.TypeOf(recordfield.Interface()).Kind() {
		case reflect.String:

			if sliceContainsString(hl7TagsList, ANNOTATION_DELIMITER) {
				// The delimiter Field itself would be pointless to cut with delimiters :)
				reflect.ValueOf(recordFieldInterface).Elem().SetString(inputFields[addr_component])
			} else if sliceContainsString(hl7TagsList, ANNOTATION_FIELDSEPARATOR) {
				reflect.ValueOf(recordFieldInterface).Elem().SetString(delimiters.Composite)
			} else if value, err := extractFieldByRepeatAndComponent(inputFields[addr_component], nextDelimiterBySubDepth, delimiters.Repeat, addr_repeat, addr_subcomponent); err == nil {
				// fmt.Printf("    %s = '%s'\n", recordfield.Type().Name(), value)
				reflect.ValueOf(recordFieldInterface).Elem().Set(reflect.ValueOf(value))
			} else {
				if inputIsRequired { // by default we ignore missing input
					return fmt.Errorf("failed to extract index (%d.%d.%d) from input line '%s' : (%w)",
						addr_component+1, addr_repeat+1, addr_subcomponent+1, inputStr, err)
				}
			}
		case reflect.Int:
			if hasOverrideDelimiterAnnotation {
				return errors.New("delimiter-annotation is only allowed for string-type, not int")
			}

			if value, err := extractFieldByRepeatAndComponent(inputFields[addr_component], nextDelimiterBySubDepth, delimiters.Repeat, addr_repeat, addr_subcomponent); err == nil {

				if num, err := strconv.Atoi(value); err == nil {
					reflect.ValueOf(recordFieldInterface).Elem().Set(reflect.ValueOf(num))
				} else {
					if inputIsRequired { // by default we ignore missing input
						return fmt.Errorf("failed to extract index (%d,%d) from field %s(%s)", addr_repeat, addr_subcomponent, inputFields[addr_component], err)
					}
				}

			} else {
				return err
			}
		case reflect.Float32:
			if hasOverrideDelimiterAnnotation {
				return fmt.Errorf("delimiter-annotation is only allowed for string-type, not int")
			}

			if value, err := extractFieldByRepeatAndComponent(inputFields[addr_component], nextDelimiterBySubDepth, delimiters.Repeat, addr_repeat, addr_subcomponent); err == nil {
				if num, err := strconv.ParseFloat(value, 32); err == nil {
					reflect.ValueOf(recordFieldInterface).Elem().Set(reflect.ValueOf(float32(num)))
				} else {
					if inputIsRequired { // by default we ignore missing input
						return fmt.Errorf("failed to extract index (%d,%d) from field %s(%w)", addr_repeat, addr_subcomponent, inputFields[addr_component], err)
					}
				}
			} else {
				return err
			}

		case reflect.Float64:
			if hasOverrideDelimiterAnnotation {
				return fmt.Errorf("delimiter-annotation is only allowed for string-type, not int")
			}

			if value, err := extractFieldByRepeatAndComponent(inputFields[addr_component], nextDelimiterBySubDepth, delimiters.Repeat, addr_repeat, addr_subcomponent); err == nil {

				if num, err := strconv.ParseFloat(value, 64); err == nil {
					reflect.ValueOf(recordFieldInterface).Elem().Set(reflect.ValueOf(float64(num)))
				} else {
					if inputIsRequired { // by default we ignore missing input
						return fmt.Errorf("failed to extract index (%d,%d) from field %s(%s)", addr_repeat, addr_subcomponent, inputFields[addr_component], err)
					}
				}

			} else {
				return err
			}
		case reflect.Struct:
			// TODO: check the package of the "Time" because it can be an another "Time" package too
			switch reflect.TypeOf(recordfield.Interface()).Name() {
			case "Time":
				if hasOverrideDelimiterAnnotation {
					return errors.New("delimiter-annotation is only allowed for string-type, not Time")
				}
				inputFieldValue := inputFields[addr_component]
				if inputFieldValue == "" {
					reflect.ValueOf(recordFieldInterface).Elem().Set(reflect.ValueOf(time.Time{}))
				} else if len(inputFieldValue) == 8 { // YYYYMMDD See Section 5.6.2 https://samson-rus.com/wp-content/files/LIS2-A2.pdf
					timeInLocation, err := time.ParseInLocation("20060102", inputFieldValue, timezone)
					if err != nil {
						return fmt.Errorf("invalid time format <%s>", inputFieldValue)
					}
					reflect.ValueOf(recordFieldInterface).Elem().Set(reflect.ValueOf(timeInLocation))

				} else if len(inputFieldValue) == 14 { // YYYYMMDDHHMMSS
					timeInLocation, err := time.ParseInLocation("20060102150405", inputFieldValue, timezone)
					if err != nil {
						return fmt.Errorf("invalid time format <%s>", inputFieldValue)
					}
					reflect.ValueOf(recordFieldInterface).Elem().Set(reflect.ValueOf(timeInLocation.UTC()))
				} else {
					return fmt.Errorf("unrecognized time format <%s>", inputFieldValue)
				}

			default:

				inputFieldValue := inputFields[addr_component]
				if err = parseStructA(inputFieldValue, subdepth+1, recordfield, timezone, delimiters); err != nil {
					return err
				}

			}

		case reflect.Slice:
			fieldParts := strings.Split(inputFields[addr_component], delimiters.Repeat)
			elementCount := len(fieldParts)
			if elementCount == 0 {
				continue
			}

			elemType := getTypeArray(recordfield.Interface())
			recordfield = reflect.MakeSlice(reflect.SliceOf(elemType), elementCount, elementCount)

			if elemType.Kind() == reflect.Struct {
				// if it is an object of the array
				for i, fieldPart := range fieldParts {
					if err = parseStructA(fieldPart, subdepth+1, recordfield.Index(i), timezone, delimiters); err != nil {
						return fmt.Errorf("unrecognized time format <%s>", fieldPart)
					}
				}
			} else {
				// its an array of prmitive types
				for i, fieldPart := range fieldParts {
					recordfield.Index(i).Set(reflect.ValueOf(fieldPart))
				}
			}
			record.Field(j).Set(recordfield)

		default:
			return fmt.Errorf("invalid field-Type '%s' for field '%s", reflect.TypeOf(recordfield.Interface()).Kind(), record.Field(j).Type().Name())
		}
	}

	return nil
}

// Translating the annotation of a field to field, index/repeat, component
// Input of one value : e.g."4" -> field -> 4
// Input of two values :"4.2" -> field, compoennt -> 4,1,2
// Input of three values "4.1.1" -> field, repeat, component -> 4,1,1
// "whereas field indexes should be 1-99 (check plz)
func readFieldAddressAnnotation(annotation string) (field int, repeat int, component int, err error) {

	if annotation == "" { // no annotation will always return the first of everything
		return 0, 0, 0, nil
	}
	field = 1
	repeat = 1
	component = 1
	fieldSplitted := strings.Split(annotation, ".")

	if len(fieldSplitted) >= 1 {
		if field, err = strconv.Atoi(fieldSplitted[0]); err != nil {
			return 0, 0, 0, err
		}
	}
	if len(fieldSplitted) >= 2 {
		if component, err = strconv.Atoi(fieldSplitted[1]); err != nil {
			return 0, 0, 0, err
		}
	}
	if len(fieldSplitted) >= 3 {
		if repeat, err = strconv.Atoi(fieldSplitted[1]); err != nil {
			return 0, 0, 0, err
		}
		if component, err = strconv.Atoi(fieldSplitted[2]); err != nil {
			return 0, 0, 0, err
		}
	}

	return field, repeat - 1, component - 1, nil
}

func getTypeArray(arr interface{}) reflect.Type {
	return reflect.TypeOf(arr).Elem()
}

// this function ettracts the field by repeat and component-delimiter
func extractFieldByRepeatAndComponent(text string, cutDelimiter, repeatDelimiter string, repeat int, component int) (string, error) {
	// fmt.Printf("      cut: %s by (%s,%s) at (%d[%d])\n", text, cutDelimiter, repeatDelimiter, repeat, component)

	subfield := strings.Split(text, repeatDelimiter)
	if repeat >= len(subfield) {
		return "", nil
	}

	subsubfield := strings.Split(subfield[repeat], cutDelimiter) // TODO:
	if component > len(subsubfield) || component < 0 {
		return "", fmt.Errorf("001 index (%d, %d) out of bounds '%s' delimiter '%s'", repeat, component, text, cutDelimiter)
	}

	if component >= len(subsubfield) {
		return "", nil
	}

	return subsubfield[component], nil
}

func sliceContainsString(list []string, search string) bool {
	for _, x := range list {
		if x == search {
			return true
		}
	}
	return false
}

func removeEmptyStrings(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}
