LIBRARY
=========

I used to share what I'm reading on Goodreads.com but over time:

- It became very slow
- It has 3 trackers as of writing this readme.
- It has many features that is confusing for me.

So I sat down and wrote my own simple library program.

- A reflection of my physical library
- It's simple
- Fast
- ~~Depends on my input (titles, images...etc)~~ : This proved to be a lot of effort to do when I first tried inserting all my books
- ~~Doesn't depend on any other system~~ : I needed the option to get the book information from google books
- Free to use
- Free to fork and modify and redistribute
- Has a feature to lend my books to other people
- Doesn't track me

# What's done so far:

- Allows adding books and taking pictures for covers from phone
- Allows creating book shelves
- Each book can be put in one shelf like real books. no multiple lists nonsense.
- User login

# Guidelines

- Keep it simple
- Minimize dependencies
- Don't add javascript
- Don't write custom CSS. use bulma.io
- ~~Don't depend on any external system like social login..etc.~~ I had to create login with google for easier implementation

# Start the server

- Clone it
- You need Go installed
- Install dependencies `go get .`
- Setup the database `bin/db setup`
- Run the server `go run *.go`

# Deployment

- a remote ssh access to a server with docker and docker-compose
- clone the repo to you machine
- copy `.env` to the remote server `/root/env/library/.env` and fill it
- from your machine `bin/deploy master user@ip-address`
- This will deploy all services to the remote server

# Contributions

- Make it simpler
- Make it faster
- Make it more secure
