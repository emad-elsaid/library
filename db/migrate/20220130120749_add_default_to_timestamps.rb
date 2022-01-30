class AddDefaultToTimestamps < ActiveRecord::Migration[6.1]
  def change
    [:books, :highlights, :shelves, :users].each do |t|
      change_column_default t, :created_at, from: nil, to: -> { 'CURRENT_TIMESTAMP' }
      change_column_default t, :updated_at, from: nil, to: -> { 'CURRENT_TIMESTAMP' }
    end
  end
end
