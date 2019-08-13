package main

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/softplan/tenkai-api/dbms/model"
	"net/http"
	"strings"
)

func commonHandler(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			return
		}

		reqToken := r.Header.Get("Authorization")
		if len(reqToken) > 0 {

			splitToken := strings.Split(reqToken, "Bearer ")
			reqToken = splitToken[1]

			token, _, errx := new(jwt.Parser).ParseUnverified(reqToken, jwt.MapClaims{})
			if errx == nil {

				if claims, ok := token.Claims.(jwt.MapClaims); ok {

					var principal model.Principal

					in := claims["realm_access"]

					realmAccessMap := in.(map[string]interface{})
					roles := realmAccessMap["roles"]
					elements := roles.([]interface{})

					for _, element := range elements {
						principal.Roles = append(principal.Roles, element.(string))
					}

					principal.Name = fmt.Sprintf("%v", claims["name"])
					principal.Email = fmt.Sprintf("%v", claims["email"])

					data, _ := json.Marshal(principal)
					r.Header.Set("principal", string(data))

					fmt.Printf("%v %v", claims["email"], claims["preferred_username"])
				} else {
					fmt.Println(errx)
				}

			} else {
				fmt.Println(errx)
			}

		}

		next.ServeHTTP(w, r)
	})
}
