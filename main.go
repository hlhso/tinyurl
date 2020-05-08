package main

func main()  {
	a:=App{}
	a.Initapp(getEnv())
	a.Run(":8000")
}
