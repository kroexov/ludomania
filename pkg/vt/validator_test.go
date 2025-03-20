package vt

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	colorGray  = "gray"
	colorBrown = "brown"

	circusNameGood = "Amazing Circus"
	circusNameBad  = "xxxxxxxxxxxxxxxxx"

	animalNameOne   = "Cecile"
	animalNameTwo   = "Shipa"
	animalNameThree = "Raymonda"
	animalNameDummy = "Dummy"
	animalNameBad   = "xxxxxxxxxxxxx"
)

type Circus struct {
	Name    string   `json:"name" validate:"required,min=3,max=16"`
	Animals []Animal `json:"animals" validate:"required,dive"`
}

type Animal struct {
	Name   string `json:"name" validate:"required,min=3,max=12"`
	Weight int    `json:"weight" validate:"required,gt=0,lt=50"`
	Tail   *Tail  `json:"optionalTail"`
}

type Tail struct {
	Color  string `json:"color" validate:"required"`
	Length int    `json:"length" validate:"gt=0"`
}

func TestValidator_CheckBasic(t *testing.T) {
	Convey("Test Validator", t, func() {
		var (
			v Validator

			tailOne   = Tail{Color: colorBrown, Length: 31}
			tailTwo   = Tail{Color: colorGray, Length: 35}
			tailThree = Tail{Color: colorBrown, Length: 28}
			tailBad   = Tail{}

			animalOne   = Animal{Name: animalNameOne, Weight: 5, Tail: &tailOne}
			animalTwo   = Animal{Name: animalNameTwo, Weight: 4, Tail: &tailTwo}
			animalThree = Animal{Name: animalNameThree, Weight: 6, Tail: &tailThree}
			animalDummy = Animal{Name: animalNameDummy, Weight: 6}
			animalBad   = Animal{Name: animalNameBad, Weight: 500, Tail: &tailBad}

			ctx = context.Background()
		)

		Convey("Positive case", func() {
			v.CheckBasic(ctx, &Circus{
				Name:    circusNameGood,
				Animals: []Animal{animalOne, animalTwo, animalThree, animalDummy},
			})
			So(v.HasErrors(), ShouldBeFalse)
			So(v.Fields(), ShouldHaveLength, 0)
		})

		Convey("Negative cases", func() {
			Convey("full", func() {
				v.CheckBasic(ctx, &Circus{
					Name:    circusNameBad,
					Animals: []Animal{animalBad},
				})
				So(v.HasErrors(), ShouldBeTrue)
				So(v.Fields(), ShouldHaveLength, 5)

				So(v.Fields(), ShouldContain, FieldError{Field: "name", Error: "max", Constraint: &FieldErrorConstraint{Max: 16}})
				So(v.Fields(), ShouldContain, FieldError{Field: "animals[0].name", Error: "max", Constraint: &FieldErrorConstraint{Max: 12}})
				So(v.Fields(), ShouldContain, FieldError{Field: "animals[0].weight", Error: "lt"})
				So(v.Fields(), ShouldContain, FieldError{Field: "animals[0].optionalTail.color", Error: "required"})
				So(v.Fields(), ShouldContain, FieldError{Field: "animals[0].optionalTail.length", Error: "required"})
			})

			Convey("good circus name", func() {
				v.CheckBasic(ctx, &Circus{
					Name:    circusNameGood,
					Animals: []Animal{animalBad},
				})
				So(v.HasErrors(), ShouldBeTrue)
				So(v.Fields(), ShouldHaveLength, 4)

				So(v.Fields(), ShouldContain, FieldError{Field: "animals[0].name", Error: "max", Constraint: &FieldErrorConstraint{Max: 12}})
				So(v.Fields(), ShouldContain, FieldError{Field: "animals[0].weight", Error: "lt"})
				So(v.Fields(), ShouldContain, FieldError{Field: "animals[0].optionalTail.color", Error: "required"})
				So(v.Fields(), ShouldContain, FieldError{Field: "animals[0].optionalTail.length", Error: "required"})
			})

			Convey("good circus name, [animalBad, animalOne]", func() {
				v.CheckBasic(ctx, &Circus{
					Name:    circusNameGood,
					Animals: []Animal{animalBad, animalOne},
				})
				So(v.HasErrors(), ShouldBeTrue)
				So(v.Fields(), ShouldHaveLength, 4)

				So(v.Fields(), ShouldContain, FieldError{Field: "animals[0].name", Error: "max", Constraint: &FieldErrorConstraint{Max: 12}})
				So(v.Fields(), ShouldContain, FieldError{Field: "animals[0].weight", Error: "lt"})
				So(v.Fields(), ShouldContain, FieldError{Field: "animals[0].optionalTail.color", Error: "required"})
				So(v.Fields(), ShouldContain, FieldError{Field: "animals[0].optionalTail.length", Error: "required"})
			})

			Convey("good circus name, [animalOne, animalBad]", func() {
				v.CheckBasic(ctx, &Circus{
					Name:    circusNameGood,
					Animals: []Animal{animalOne, animalBad},
				})
				So(v.HasErrors(), ShouldBeTrue)
				So(v.Fields(), ShouldHaveLength, 4)

				So(v.Fields(), ShouldContain, FieldError{Field: "animals[1].name", Error: "max", Constraint: &FieldErrorConstraint{Max: 12}})
				So(v.Fields(), ShouldContain, FieldError{Field: "animals[1].weight", Error: "lt"})
				So(v.Fields(), ShouldContain, FieldError{Field: "animals[1].optionalTail.color", Error: "required"})
				So(v.Fields(), ShouldContain, FieldError{Field: "animals[1].optionalTail.length", Error: "required"})
			})

			Convey("good circus name, [animalOne, animalBad, animalDummy]", func() {
				v.CheckBasic(ctx, &Circus{
					Name:    circusNameGood,
					Animals: []Animal{animalOne, animalBad, animalDummy},
				})
				So(v.HasErrors(), ShouldBeTrue)
				So(v.Fields(), ShouldHaveLength, 4)

				So(v.Fields(), ShouldContain, FieldError{Field: "animals[1].name", Error: "max", Constraint: &FieldErrorConstraint{Max: 12}})
				So(v.Fields(), ShouldContain, FieldError{Field: "animals[1].weight", Error: "lt"})
				So(v.Fields(), ShouldContain, FieldError{Field: "animals[1].optionalTail.color", Error: "required"})
				So(v.Fields(), ShouldContain, FieldError{Field: "animals[1].optionalTail.length", Error: "required"})
			})

			Convey("good circus name, [animalOne, animalBad, animalTwo, animalBad]", func() {
				v.CheckBasic(ctx, &Circus{
					Name:    circusNameGood,
					Animals: []Animal{animalOne, animalBad, animalTwo, animalBad},
				})
				So(v.HasErrors(), ShouldBeTrue)
				So(v.Fields(), ShouldHaveLength, 8)

				So(v.Fields(), ShouldContain, FieldError{Field: "animals[1].name", Error: "max", Constraint: &FieldErrorConstraint{Max: 12}})
				So(v.Fields(), ShouldContain, FieldError{Field: "animals[1].weight", Error: "lt"})
				So(v.Fields(), ShouldContain, FieldError{Field: "animals[1].optionalTail.color", Error: "required"})
				So(v.Fields(), ShouldContain, FieldError{Field: "animals[1].optionalTail.length", Error: "required"})

				So(v.Fields(), ShouldContain, FieldError{Field: "animals[3].name", Error: "max", Constraint: &FieldErrorConstraint{Max: 12}})
				So(v.Fields(), ShouldContain, FieldError{Field: "animals[3].weight", Error: "lt"})
				So(v.Fields(), ShouldContain, FieldError{Field: "animals[3].optionalTail.color", Error: "required"})
				So(v.Fields(), ShouldContain, FieldError{Field: "animals[3].optionalTail.length", Error: "required"})
			})

			Convey("good circus name, no animals", func() {
				v.CheckBasic(ctx, &Circus{
					Name: circusNameGood,
				})
				So(v.HasErrors(), ShouldBeTrue)
				So(v.Fields(), ShouldHaveLength, 1)
				So(v.Fields(), ShouldContain, FieldError{Field: "animals", Error: "required"})
			})

			Convey("good circus name, one overweighted animal", func() {
				v.CheckBasic(ctx, &Circus{
					Name:    circusNameGood,
					Animals: []Animal{{Name: animalNameOne, Weight: 200}},
				})
				So(v.HasErrors(), ShouldBeTrue)
				So(v.Fields(), ShouldHaveLength, 1)
				So(v.Fields(), ShouldContain, FieldError{Field: "animals[0].weight", Error: "lt"})
			})
		})
	})
}
