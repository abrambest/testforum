package internal

import


func createDB(){
	db, err := sql.Open("mysql3", "./data")
	if err != nil{
		fmt.Println(err)
	}
}