# ocdSearch

Start of a search interface based on the bleve template package.

## What

A search of the JSON-LD index for OCD

## Docker notes

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go
docker build --tag="opencoredata/ocdsearch:0.2" .

## Examples

## Indexing 
Some notes on indexing.  currently I have five indexes.

* abstracts (CSDCO abstracts)
* compositIndex (JRSO and CSDCO resources indexed)
* CSDCO  (resource index for CSDCO)
* JRSO  (resource index for JRSO)
* test (a test index for some, well, testing...)

Really the compositIndex is redundant and remove some flexibility in searching.  So I should just make the 
three indexs of: JRSO, CSDCO and abstracts.  


## Future work

- Get it working :)
- Use iron-form https://elements.polymer-project.org/elements/iron-form?view=demo:demo/index.html&active=iron-form

