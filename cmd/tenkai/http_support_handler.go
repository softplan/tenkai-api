package main

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
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
					r.Header.Set("principal", fmt.Sprintf("%v", claims["email"]))

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




