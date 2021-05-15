class Book < ActiveRecord::Base
  belongs_to :shelf
  belongs_to :user, required: true

  validates_presence_of :title, :author, :isbn

  default_scope { order(created_at: :desc) }
  before_destroy { File.delete "public/books/image/#{image}" if image }
end

class Shelf < ActiveRecord::Base
  belongs_to :user, required: true
  has_many :books, dependent: :nullify

  validates_presence_of :name
end

class User < ActiveRecord::Base
  has_many :books, dependent: :destroy
  has_many :shelves, dependent: :destroy

  validates :email, presence: true, uniqueness: true
end
