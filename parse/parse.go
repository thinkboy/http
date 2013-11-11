package parse

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

type tagInfo struct {
	Name      string
	OmitEmpty bool
	RangeStr  string
}

/*
	note :
		1.make sure the params of url.Values are lowercase
		2.you need to assign a pointer of struct to the second parameter
*/

func ParseUrlParam(values url.Values, stru interface{}) error {
	sv := reflect.ValueOf(stru).Elem()
	st := sv.Type()
	n := sv.NumField()

	for i := 0; i < n; i++ {
		stField := st.Field(i)
		svField := sv.Field(i)

		tag := stField.Tag.Get("param")
		if tag == "-" {
			continue
		}

		//fmt.Printf("field \"%s\", tag: \"%s\"\n", stField.Name, tag)

		// Get tags
		ti := tagInfo{}
		ti.Name = stField.Name
		if tag != "" {
			tags := strings.Split(tag, ",")
			for _, oneTag := range tags {
				if oneTag == "omitempty" {
					ti.OmitEmpty = true

				} else if strings.HasPrefix(oneTag, "range") {
					// exmple : range[32] or range[:9]
					lenTagStr := len(oneTag)

					if lenTagStr > 7 && oneTag[5:6] == "[" && oneTag[lenTagStr-1:lenTagStr] == "]" {
						ti.RangeStr = oneTag[6 : lenTagStr-1]
					} else {
						return errors.New(fmt.Sprintf("invalid tag \"%s\"", oneTag))
					}

				} else {
					return errors.New(fmt.Sprintf("invalid tag \"%s\"", oneTag))
				}
			}
		}

		// Find values
		value := values.Get(strings.ToLower(stField.Name))

		//fmt.Printf("field \"%s\", value: \"%s\"\n", stField.Name, value)

		switch svField.Kind() {
		case reflect.String:
			if value == "" {
				if ti.OmitEmpty == false {
					return errors.New(fmt.Sprintf("field \"%s\" is empty", stField.Name))
				}
			} else {
				if ti.RangeStr != "" {
					num, err := strconv.Atoi(ti.RangeStr)
					if err != nil {
						return errors.New(fmt.Sprintf("invalid tag range \"%s\" error: %v", ti.RangeStr, err))
					}

					if len(value) != num {
						return errors.New(fmt.Sprintf("field \"%s\" is not %d character", stField.Name, num))
					}
				}
			}

			svField.SetString(value)
		case reflect.Slice:
			var (
				data []byte
				e    error
			)

			if value == "" {
				if ti.OmitEmpty == false {
					return errors.New(fmt.Sprintf("field \"%s\" is empty", stField.Name))
				}
			} else {
				if ti.RangeStr != "" {
					num, err := strconv.Atoi(ti.RangeStr)
					if err != nil {
						return errors.New(fmt.Sprintf("invalid tag range \"%s\" error: %v", ti.RangeStr, err))
					}

					data, e = hex.DecodeString(value)
					if e != nil {
						return errors.New(fmt.Sprintf("invalid tag \"%s\" error: %v", ti.RangeStr, err))
					}

					if len(data) != num {
						return errors.New(fmt.Sprintf("field \"%s\" is not %d bytes", stField.Name, num))
					}
				}
			}

			svField.SetBytes(data)
		case reflect.Int:
			if value == "" {
				if ti.OmitEmpty == false {
					return errors.New(fmt.Sprintf("field \"%s\" is empty", stField.Name))
				}
			} else {
				val, err := strconv.Atoi(value)
				if err != nil {
					return errors.New(fmt.Sprintf("invalid int value \"%s\" error: %v", value, err))
				}

				// Verify range
				if ti.RangeStr != "" {
					index := strings.Index(ti.RangeStr, ":")
					if index == -1 {
						return errors.New(fmt.Sprintf("invalid range tag \"%s\"", ti.RangeStr))
					}

					minStr := ti.RangeStr[:index]

					if minStr != "" {
						min, err := strconv.Atoi(minStr)
						if err != nil {
							return errors.New(fmt.Sprintf("invalid tag range \"%s\" error: %v", ti.RangeStr, err))
						}

						if val < min {
							return errors.New(fmt.Sprintf("field \"%s\" out of range", stField.Name))
						}
					}

					maxStr := ti.RangeStr[index+1:]

					if maxStr != "" {
						max, err := strconv.Atoi(maxStr)
						if err != nil {
							return errors.New(fmt.Sprintf("invalid tag range \"%s\" error: %v", ti.RangeStr, err))
						}

						if val > max {
							return errors.New(fmt.Sprintf("field \"%s\" out of range", stField.Name))
						}
					}

				}

				// It is converted to 32 bits inside SetInt()
				svField.SetInt(int64(val))
			}

		case reflect.Int64:
			if value == "" {
				if ti.OmitEmpty == false {
					return errors.New(fmt.Sprintf("field \"%s\" is empty", stField.Name))
				}
			} else {
				val, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					return errors.New(fmt.Sprintf("invalid int value \"%s\" error: %v", value, err))
				}

				// Verify range
				if ti.RangeStr != "" {
					index := strings.Index(ti.RangeStr, ":")
					if index == -1 {
						return errors.New(fmt.Sprintf("invalid range tag \"%s\"", ti.RangeStr))
					}

					minStr := ti.RangeStr[:index]

					if minStr != "" {
						min, err := strconv.ParseInt(minStr, 10, 64)
						if err != nil {
							return errors.New(fmt.Sprintf("invalid tag range \"%s\" error: %v", ti.RangeStr, err))
						}

						if val < min {
							return errors.New(fmt.Sprintf("field \"%s\" out of range", stField.Name))
						}
					}

					maxStr := ti.RangeStr[index+1:]

					if maxStr != "" {
						max, err := strconv.ParseInt(maxStr, 10, 64)
						if err != nil {
							return errors.New(fmt.Sprintf("invalid tag range \"%s\" error: %v", ti.RangeStr, err))
						}

						if val > max {
							return errors.New(fmt.Sprintf("field \"%s\" out of range", stField.Name))
						}
					}

				}

				svField.SetInt(val)
			}
			//TODO:need to add more type later
			//case reflect.Float32:
			//case reflect.Float64:
		default:
			return errors.New(fmt.Sprintf("field \"%s\" no this type \"%s\" deal", stField.Name))

		}
	}

	return nil
}

func ParseBodyParam(rBody io.ReadCloser, stru interface{}) (body []byte, erro error) {
	sv := reflect.ValueOf(stru).Elem()
	st := sv.Type()
	n := sv.NumField()

	values, tmpBody, err := getPostParams(rBody)
	body = tmpBody
	if err != nil {
		erro = err
		return
	}

	for i := 0; i < n; i++ {
		stField := st.Field(i)
		svField := sv.Field(i)

		tag := stField.Tag.Get("param")
		if tag == "-" {
			continue
		}

		//fmt.Printf("field \"%s\", tag: \"%s\"\n", stField.Name, tag)

		// Get tags
		ti := tagInfo{}
		ti.Name = stField.Name
		if tag != "" {
			tags := strings.Split(tag, ",")
			for _, oneTag := range tags {
				if oneTag == "omitempty" {
					ti.OmitEmpty = true

				} else if strings.HasPrefix(oneTag, "range") {
					// exmple : range[32] or range[:9]
					lenTagStr := len(oneTag)

					if lenTagStr > 7 && oneTag[5:6] == "[" && oneTag[lenTagStr-1:lenTagStr] == "]" {
						ti.RangeStr = oneTag[6 : lenTagStr-1]
					} else {
						erro = errors.New(fmt.Sprintf("invalid tag \"%s\"", oneTag))
						return
					}

				} else {
					erro = errors.New(fmt.Sprintf("invalid tag \"%s\"", oneTag))
					return
				}
			}
		}

		// Find values
		value, _ := values[strings.ToLower(stField.Name)]

		//fmt.Printf("field \"%s\", value: \"%s\"\n", stField.Name, value)

		switch svField.Kind() {
		case reflect.String:
			if value == "" {
				if ti.OmitEmpty == false {
					erro = errors.New(fmt.Sprintf("field \"%s\" is empty", stField.Name))
					return
				}
			} else {
				if ti.RangeStr != "" {
					num, err := strconv.Atoi(ti.RangeStr)
					if err != nil {
						erro = errors.New(fmt.Sprintf("invalid tag range \"%s\" error: %v", ti.RangeStr, err))
						return
					}

					if len(value) != num {
						erro = errors.New(fmt.Sprintf("field \"%s\" is not %d character", stField.Name, num))
						return
					}
				}
			}

			svField.SetString(value)
		case reflect.Slice:
			var (
				data []byte
				e    error
			)

			if value == "" {
				if ti.OmitEmpty == false {
					erro = errors.New(fmt.Sprintf("field \"%s\" is empty", stField.Name))
					return
				}
			} else {
				if ti.RangeStr != "" {
					num, err := strconv.Atoi(ti.RangeStr)
					if err != nil {
						erro = errors.New(fmt.Sprintf("invalid tag range \"%s\" error: %v", ti.RangeStr, err))
						return
					}

					data, e = hex.DecodeString(value)
					if e != nil {
						erro = errors.New(fmt.Sprintf("invalid tag \"%s\" error: %v", ti.RangeStr, err))
						return
					}

					if len(data) != num {
						erro = errors.New(fmt.Sprintf("field \"%s\" is not %d bytes", stField.Name, num))
						return
					}
				}
			}

			svField.SetBytes(data)
		case reflect.Int:
			if value == "" {
				if ti.OmitEmpty == false {
					erro = errors.New(fmt.Sprintf("field \"%s\" is empty", stField.Name))
					return
				}
			} else {
				val, err := strconv.Atoi(value)
				if err != nil {
					erro = errors.New(fmt.Sprintf("invalid int value \"%s\" error: %v", value, err))
					return
				}

				// Verify range
				if ti.RangeStr != "" {
					index := strings.Index(ti.RangeStr, ":")
					if index == -1 {
						erro = errors.New(fmt.Sprintf("invalid range tag \"%s\"", ti.RangeStr))
						return
					}

					minStr := ti.RangeStr[:index]

					if minStr != "" {
						min, err := strconv.Atoi(minStr)
						if err != nil {
							erro = errors.New(fmt.Sprintf("invalid tag range \"%s\" error: %v", ti.RangeStr, err))
							return
						}

						if val < min {
							erro = errors.New(fmt.Sprintf("field \"%s\" out of range", stField.Name))
							return
						}
					}

					maxStr := ti.RangeStr[index+1:]

					if maxStr != "" {
						max, err := strconv.Atoi(maxStr)
						if err != nil {
							erro = errors.New(fmt.Sprintf("invalid tag range \"%s\" error: %v", ti.RangeStr, err))
							return
						}

						if val > max {
							erro = errors.New(fmt.Sprintf("field \"%s\" out of range", stField.Name))
							return
						}
					}

				}

				// It is converted to 32 bits inside SetInt()
				svField.SetInt(int64(val))
			}

		case reflect.Int64:
			if value == "" {
				if ti.OmitEmpty == false {
					erro = errors.New(fmt.Sprintf("field \"%s\" is empty", stField.Name))
					return
				}
			} else {
				val, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					erro = errors.New(fmt.Sprintf("invalid int value \"%s\" error: %v", value, err))
					return
				}

				// Verify range
				if ti.RangeStr != "" {
					index := strings.Index(ti.RangeStr, ":")
					if index == -1 {
						erro = errors.New(fmt.Sprintf("invalid range tag \"%s\"", ti.RangeStr))
						return
					}

					minStr := ti.RangeStr[:index]

					if minStr != "" {
						min, err := strconv.ParseInt(minStr, 10, 64)
						if err != nil {
							erro = errors.New(fmt.Sprintf("invalid tag range \"%s\" error: %v", ti.RangeStr, err))
							return
						}

						if val < min {
							erro = errors.New(fmt.Sprintf("field \"%s\" out of range", stField.Name))
							return
						}
					}

					maxStr := ti.RangeStr[index+1:]

					if maxStr != "" {
						max, err := strconv.ParseInt(maxStr, 10, 64)
						if err != nil {
							erro = errors.New(fmt.Sprintf("invalid tag range \"%s\" error: %v", ti.RangeStr, err))
							return
						}

						if val > max {
							erro = errors.New(fmt.Sprintf("field \"%s\" out of range", stField.Name))
							return
						}
					}

				}

				svField.SetInt(val)
			}
			//TODO:need to add more type later
			//case reflect.Float32:
			//case reflect.Float64:
		default:
			erro = errors.New(fmt.Sprintf("field \"%s\" no this type \"%s\" deal", stField.Name))
			return

		}
	}

	return
}

func getPostParams(r io.ReadCloser) (map[string]string, []byte, error) {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, nil, err
	}
	defer r.Close()

	params := make(map[string]string)
	bodyTmp, err := url.QueryUnescape(string(body))
	if err != nil {
		return nil, body, err
	}

	postParams := strings.Split(string(bodyTmp), "&")
	for _, postParam := range postParams {
		index := strings.Index(postParam, "=")
		if index == -1 {
			continue
		}
		key := postParam[:index]
		param := postParam[index+1:]
		params[key] = param
	}

	return params, body, nil
}
