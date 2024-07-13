# Reddit Viewer
This was a small exercise in web scraping and re-rendering. This code is not
suitable for production use, but it does demonstrate how the contents of a
reasonably static web page can be parsed into some domain model that can then
be manipulated programmatically or re-exported in some other format.

The application hosts a local HTTP server on `http://localhost:8080` that 
effectively proxies between the user and (old) Reddit. If the user navigates to
`http://localhost:8080/r/comics`, the server will download the corresponding 
page from Reddit (http://old.reddit.com/r/comics) and extract a model from the 
page contents using HTTP scraping techniques. The model represents an 
abstract representation of the page (e.g. a collection of posts, where each 
post has a title, a timestamp, a score, etc.).

The model can easily be exported in a JSON format to allow for data analysis 
or interaction with other tools. We also show how the model can be used as a 
data source for a "re-rendering" pass. We create custom HTTP templates and 
instantiate them using the data provided in the model, allowing for dynamic
customization of the presentation of the data.

Ultimately, this project is primarily an academic exercise, since Reddit does
expose a JSON API directly and neither the model we create nor the re-rendered 
HTML pages would be considered "complete" in the sense that they cover all of
Reddit, but it does demonstrate how some of those concepts can be achieved.

## Developer Information

Start the local server. The server will start listening on `http://localhost:8080`.
```bash
go run main.go
```
