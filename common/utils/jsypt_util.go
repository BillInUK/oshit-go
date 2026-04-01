package utils

import (
	"errors"
	"fmt"
	"github.com/Mystery00/go-jasypt"
	"github.com/Mystery00/go-jasypt/iv"
	"github.com/Mystery00/go-jasypt/salt"
	"reflect"
	"strings"
)

const JasyptDefaultAlgorithm = "PBEWithHMACSHA512AndAES_256"

func JasyptDecode(obj interface{}, key, algorithm string) error {
	var v reflect.Value
	if ov, ok := obj.(reflect.Value); ok {
		v = ov
	} else {
		v = reflect.ValueOf(obj)
	}
	switch v.Kind() {
	case reflect.Ptr:
		return JasyptDecode(v.Elem(), key, algorithm)
	case reflect.String:
		str := v.String()
		if v.CanSet() && strings.HasPrefix(str, "ENC~") {
			text, err := JasyptDecrypt(str[4:], key, algorithm)
			if err != nil {
				return err
			}
			v.SetString(string(text))
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			if err := JasyptDecode(field, key, algorithm); err != nil {
				return err
			}
		}
	case reflect.Slice | reflect.Array:
		l := v.Len()
		for i := 0; i < l; i++ {
			if err := JasyptDecode(v.Index(i), key, algorithm); err != nil {
				return err
			}
		}
	case reflect.Interface:
		return JasyptDecode(v.Interface(), key, algorithm)
	}
	return nil
}

func JasyptEncrypt(message string, password string, algorithm string) (string, error) {
	if len(algorithm) == 0 {
		algorithm = JasyptDefaultAlgorithm
	} else if !isValidAlgorithm(algorithm) {
		return "", errors.New(fmt.Sprintf("Unsupported algorithm: %s", algorithm))
	}

	// create a new instance of jasypt
	encryptor := jasypt.New(algorithm, jasypt.NewConfig(
		jasypt.SetPassword(password),
		jasypt.SetSaltGenerator(salt.RandomSaltGenerator{}),
		jasypt.SetIvGenerator(iv.RandomIvGenerator{}),
	))

	// encrypt the message
	return encryptor.Encrypt(message)
}

func JasyptDecrypt(encode string, password string, algorithm string) (string, error) {
	if len(algorithm) == 0 {
		algorithm = JasyptDefaultAlgorithm
	} else if !isValidAlgorithm(algorithm) {
		return "", errors.New(fmt.Sprintf("Unsupported algorithm: %s", algorithm))
	}

	// create a new instance of jasypt
	encryptor := jasypt.New(algorithm, jasypt.NewConfig(
		jasypt.SetPassword(password),
		jasypt.SetSaltGenerator(salt.RandomSaltGenerator{}),
		jasypt.SetIvGenerator(iv.RandomIvGenerator{}),
	))
	// decrypt the message
	return encryptor.Decrypt(encode)
}

func isValidAlgorithm(algorithm string) bool {
	for _, currentAlgo := range supportedAlgorithms() {
		if currentAlgo == algorithm {
			return true
		}
	}
	return false
}

func supportedAlgorithms() []string {
	return []string{"PBEWithHMACSHA512AndAES_256", "PBEWithMD5AndDES"}
}
