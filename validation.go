package main

func (n NewBookParams) Validate() ValidationErrors {
	ve := ValidationErrors{}
	ValidateStringPresent(n.Title, "title", "Title", ve)
	ValidateStringLength(n.Title, "title", "Title", ve, 0, 100)

	ValidateStringLength(n.Subtitle, "subtitle", "Subtitle", ve, 0, 100)

	ValidateStringPresent(n.Author, "author", "Author", ve)
	ValidateStringLength(n.Author, "author", "Author", ve, 0, 100)

	ValidateStringNumeric(n.Isbn, "isbn", "ISBN", ve)
	ValidateISBN13(n.Isbn, "isbn", "ISBN", ve)

	ValidateStringLength(n.GoogleBooksID.String, "google_books_id", "Google Books ID", ve, 0, 30)
	ValidateStringLength(n.Description, "description", "Description", ve, 0, 2000)
	ValidateStringLength(n.Publisher, "publisher", "Publisher", ve, 0, 50)

	return ve
}

func (n UpdateBookParams) Validate() ValidationErrors {
	ve := ValidationErrors{}
	ValidateStringPresent(n.Title, "title", "Title", ve)
	ValidateStringLength(n.Title, "title", "Title", ve, 0, 100)

	ValidateStringLength(n.Subtitle, "subtitle", "Subtitle", ve, 0, 100)

	ValidateStringPresent(n.Author, "author", "Author", ve)
	ValidateStringLength(n.Author, "author", "Author", ve, 0, 100)

	ValidateStringLength(n.Description, "description", "Description", ve, 0, 2000)
	ValidateStringLength(n.Publisher, "publisher", "Publisher", ve, 0, 50)
	return ve
}

func (u UpdateUserParams) Validate() ValidationErrors {
	ve := ValidationErrors{}
	ValidateStringLength(u.Description.String, "description", "Description", ve, 0, 500)
	ValidateStringLength(u.AmazonAssociatesID.String, "amazon_associates_id", "Amazon Associates ID", ve, 0, 50)
	ValidateStringLength(u.Facebook.String, "facebook", "Facebook", ve, 0, 50)
	ValidateStringLength(u.Twitter.String, "twitter", "Twitter", ve, 0, 50)
	ValidateStringLength(u.Linkedin.String, "linkedin", "Linkedin", ve, 0, 50)
	ValidateStringLength(u.Instagram.String, "instagram", "Instagram", ve, 0, 50)
	ValidateStringLength(u.Phone.String, "phone", "Phone", ve, 0, 50)
	ValidateStringLength(u.Whatsapp.String, "whatsapp", "Whatsapp", ve, 0, 50)
	ValidateStringLength(u.Telegram.String, "telegram", "Telegram", ve, 0, 50)
	return ve
}

func (n NewHighlightParams) Validate() ValidationErrors {
	ve := ValidationErrors{}
	ValidateStringLength(n.Content, "content", "Content", ve, 10, 500)
	ValidateInt32Min(n.Page, "page", "Page", ve, 0)
	return ve
}

func (n UpdateHighlightParams) Validate() ValidationErrors {
	ve := ValidationErrors{}
	ValidateStringLength(n.Content, "content", "Content", ve, 10, 500)
	ValidateInt32Min(n.Page, "page", "Page", ve, 0)
	return ve
}
