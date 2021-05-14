class Book < ActiveRecord::Base
  validates_presence_of :title, :author, :isbn
  belongs_to :shelf
  default_scope { order(created_at: :desc) }
  before_destroy { File.delete "public/books/image/#{image}" if image }
end

class Shelf < ActiveRecord::Base
  validates_presence_of :name
  has_many :books, dependent: :nullify
end

class User < ActiveRecord::Base
  validates :email, presence: true, uniqueness: true
end
