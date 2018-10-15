# ServersAccess
The intent of this tool is to simplayfy the access(ssh/winscp/http) of any servers in infrastacture. 

#Achnoledgement:
I want to extend my thanks to Putty (https://www.putty.org/), Winscp(https://winscp.net) and sqllite (https://www.sqlite.org) for creating an awsome tools.
ServerAccess tool uses sqllite to store servers information. It send request installed Putty and winscp installed exe on, local machine to invoke connection.

#How to use:
1. install putty and windscp.
2. download sqllite and place it in same folder as main.go.
3. Run main.go 
4. access http://localhost:8080
