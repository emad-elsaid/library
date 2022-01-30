class BookUserIdNotNull < ActiveRecord::Migration[6.1]
  def change
    change_column_null :books, :user_id, false, 1
    change_column_null :books, :title, false, ""
    change_column_null :books, :author, false, ""
    change_column_null :books, :isbn, false, ""
    change_column_null :books, :subtitle, false, ""
    change_column_null :books, :description, false, ""
    change_column_null :books, :page_count, false,  0
    change_column_null :books, :publisher, false,  ""
  end
end
