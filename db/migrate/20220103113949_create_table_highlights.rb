class CreateTableHighlights < ActiveRecord::Migration[6.1]
  def change
    create_table :highlights do |t|
      t.belongs_to :book, foreign_key: true, null: false
      t.integer :page, null: false
      t.string :content, null: false
      t.string :image

      t.timestamps
    end
  end
end
