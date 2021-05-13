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
- Depends on my input (titles, images...etc)
- Doesn't depend on any other system
- Free to use
- Free to fork and modify and redistribute
- Has a feature to lend my books to other people

What's done and what's not:

- So far it's single user. for local use
- Allows adding books and taking pictures for covers from phones.
- Allows creating book shelves
- Each book can be put in one shelf like real books. no multiple lists nonsense.

Guideline:

- Keep it simple
- Minimize dependencies
- Don't add javascript
- Don't write custom CSS. use bulma.io
- Don't depend on any external system like social login..etc.

Start the server

- Clone it
- Install dependencies `bundle install`
- Run the server `bundle exec ./main`
