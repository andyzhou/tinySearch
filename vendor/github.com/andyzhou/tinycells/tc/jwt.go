package tc

import (
	"github.com/dgrijalva/jwt-go"
	"log"
)

/*
 * Internal JWT interface
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//jwt info
type JWT struct {
	secret string `secret key string`
	token *jwt.Token `jwt token instance`
	claims jwt.MapClaims `jwt claims object`
}

//construct
func NewJWT(secretKey string) *JWT {
	this := &JWT{
		secret:secretKey,
		token:jwt.New(jwt.SigningMethodHS256),
		claims:make(jwt.MapClaims),
	}
	return this
}

//encode
func (j *JWT) Encode(input map[string]interface{}) string {
	j.claims = input
	j.token.Claims = j.claims
	result, err := j.token.SignedString([]byte(j.secret))
	if err != nil {
		log.Println("Encode jwt failed, error:", err.Error())
		return ""
	}
	return result
}

//decode
func (j *JWT) Decode(input string) map[string]interface{} {
	//parse input string
	token, err := jwt.Parse(input, j.getValidationKey)
	if err != nil {
		log.Println("Decode ", input, " failed, error:", err.Error())
		return nil
	}
	//check header
	if jwt.SigningMethodHS256.Alg() != token.Header["alg"] {
		log.Println("Header error")
		return nil
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims
	}
	return nil
}

//get validate key
func (j *JWT) getValidationKey(*jwt.Token) (interface{}, error) {
	return []byte(j.secret), nil
}
