# go-hl7 ![build status](https://travis-ci.org/78bit/uuid.svg?branch=master)

Golang library for handling hl7 2.x Procotol

###### Install
`go get github.com/DRK-Blutspende-BaWueHe/go-hl7`

## Features
  - Encoding
    - UTF8
    - ASCII
    - Windows1250
    - Windows1251
    - Windows1252
    - DOS852
    - DOS855
    - DOS866
    - ISO8859_1
  - Timezone Support
  - Marshal/Unmarshal function

## Quick Start

The following Go code decodes a hl7 read from a File.

``` go
fileData, err := ioutil.ReadFile("tbd.hl7")
if err != nil {
    log.Fatal(err)
}

var message 
err := hl7.Unmarshal(
    []byte(filedata),
    &message,
    hl7.EncodingUTF8,
    hl7.TimezoneEuropeBerlin)

if err != nil {
   log.Fatal(err)
}
```

The following Go code encodes a hl7 message.

``` go
marshalledMessageBytes, err := hl7.Marshal(
    message,
    hl7.StandardFieldSeparator,
    hl7.EncodingASCII,
    hl7.TimezoneEuropeBerlin,
    hl7.StandardNotation)

if err != nil {
   log.Fatal(err)
}
```

To identify the version and type of the message, you can use the IdentifyMessage function.

```go
var data string
messageType, protocolVersion, err := hl7.IdentifyMessage(
    []byte(data),
    hl7.EncodingUTF8,
)
```

## Custom Messages and Segments

### Segment Definition

You can define custom segments with the hl7 annotation. Field #0 is the Segment-identifier itself, here AL1. 

```golang
// HL7 v2.4 - AL1 - Patient allergy information
// https://hl7-definition.caristix.com/v2/HL7v2.4/Segments/AL1
type AL1 struct {
    SetId               CE        `hl7:"1"`
    AllergenTypeCode    CE        `hl7:"2"`
    AllergenCode        CE        `hl7:"3"`
    AllergySeverityCode CE        `hl7:"4"`
    AllergyReactionCode string    `hl7:"5"`
    IdentificationDate  time.Time `hl7:"6,shortdate"`
}
``` 

### Message Defintion

Messages are an arrangement of Segments. Use struct and []struct to group the records like so. You can then use these messages as target in the unmarshal-operation:

```golang
// HL7 v2.4 - SSU_U03 - Specimen status update
// https://hl7-definition.caristix.com/v2/HL7v2.4/TriggerEvents/SSU_U03
type SSU_U03 struct {
    MSH               MSH `hl7:"MSH" json:"MSH"`
    EquipmentDetail   EQU `hl7:"EQU" json:"EquipmentDetail"`
    SpecimenContainer []struct {
        SpecimenContainerDetail SAC `hl7:"SAC" json:"SpecimenContainerDetail"`
        ObservationResult       OBX `hl7:"OBX" json:"ObservationResult"`
    }
    Role ROL `hl7:"ROL,optional" json:"Role"`
}
```
