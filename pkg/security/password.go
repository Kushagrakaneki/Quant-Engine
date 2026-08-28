package security

import "golang.org/x/crypto/bcrypt"

//takes og password and converts it into hash version(with salt and cost) which can only be used for verification later when user logs in
func HashPassword(password string)(string,error){
	
	bytes,err:=bcrypt.GenerateFromPassword([]byte(password),14)

	return string(bytes),err
}

//takes entered password and hashes it using same salt and cost(stored inside hash itself) and then it checks with stored hash 
func CheckPassword(password string,hash string) bool{
	err:=bcrypt.CompareHashAndPassword([]byte(hash),[]byte(password))//order matters of argument

	return err==nil
}