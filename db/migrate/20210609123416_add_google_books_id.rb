class AddGoogleBooksId < ActiveRecord::Migration[6.1]
  def change
    add_column :books, :google_books_id, :string
  end
end
