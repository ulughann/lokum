package lokum

import (
	"errors"
	"fmt"
)

var (
	ErrStackOverflow = errors.New("stack overflow")

	ErrObjectAllocLimit = errors.New("obje limiti aşıldı")

	ErrIndexOutOfBounds = errors.New("dizin sınır dışı")

	ErrInvalidIndexType = errors.New("geçersiz dizin")

	ErrInvalidIndexValueType = errors.New("geçersiz dizin")

	ErrInvalidIndexOnError = errors.New("hatada geçersiz dizin")

	ErrInvalidOperator = errors.New("geçersiz operatör")

	ErrWrongNumArguments = errors.New("yanlış argüman sayısı girilmiş")

	ErrBytesLimit = errors.New("byte boyut limiti geçildi")

	ErrStringLimit = errors.New("string boyut limiti geçildi")

	ErrNotIndexable = errors.New("index alınamaz")

	ErrNotIndexAssignable = errors.New("index atanamaz")

	ErrNotImplemented = errors.New("henüz uygulanmadı")

	ErrInvalidRangeStep = errors.New("range 0dan büyük olmalı")
)

type ErrInvalidArgumentType struct {
	Name     string
	Expected string
	Found    string
}

func (e ErrInvalidArgumentType) Error() string {
	return fmt.Sprintf("argüman '%s' için geçersiz tip. %s beklendi, %s bulundu",
		e.Name, e.Expected, e.Found)
}
