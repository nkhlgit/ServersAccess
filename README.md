# ServersAccess
The intent of this tool is to simplayfy the access(ssh/winscp/http) of any servers in infrastacture. 

## Acknowledgement
I want to extend my thanks to Putty (https://www.putty.org/), Winscp(https://winscp.net) and sqllite (https://www.sqlite.org) for creating an awsome tools.
ServerAccess tool uses sqllite to store servers information. It send request installed Putty and winscp installed exe on, local machine to invoke connection.

## How to use
1. install putty and windscp.
2. download sqllite and place it in same folder as main.go.
3. Run main.go 
4. access http://localhost:8080
5. For bulk import of server data, please refer the template infra_template.csv.<br />
  feild : description<br />
  name = Unique name of the serer. I can any string.<br />
  ip = Serever ip address.<br />
  hostname = server hostname
  osUser = User access of OS. this user will be used for ssh and RDP access
  osPassword = osUser Password
  osPort = port open for SSH / RDP
  webPort = port open for we access
  product = Product for user refrence
  datacenter = Name of data centter for user refrence.
  webPrefix = web prefix like http or https.
  webSuffix = web-suiiffx. anyting after htt://server-IP:Port/
  fav = (y/n) reserfed for futhure use to add faviourate servers.
