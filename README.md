LIBRARY
=========

I used share what I'm reading on Goodreads.com but over time:

- It became very slow
- It has 3 trackers as of writing this readme.
- It has many features that is confusing for me.
- It doesn't offer a way to lend books to my friends

So I sat down and wrote my own simple library program.

- A reflection on my physical library
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
- Install dependencies `bundle install`
- Setup the database `rake db:setup`
- Run the server `bundle exec ./main`

# Sending emails

- The `mail` service/script will run every hour
- Emails are grouped into single message
- `mail` uses ruby standard library SMTP module using the configuration in `.env`

# Contributions

- Make it simpler
- Make faster
- Make it more secure
