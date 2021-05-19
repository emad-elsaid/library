class Book < ActiveRecord::Base
  belongs_to :shelf
  belongs_to :user, required: true
  has_many :borrows

  validates_presence_of :title, :author, :isbn

  default_scope { order(created_at: :desc) }
  before_destroy { File.delete "public/books/image/#{image}" if image }
end

class Shelf < ActiveRecord::Base
  belongs_to :user, required: true
  has_many :books, dependent: :nullify
  acts_as_list scope: :user

  validates_presence_of :name
end

class Borrow < ActiveRecord::Base
  belongs_to :user, required: true, inverse_of: :borrows
  belongs_to :owner, required: true, class_name: :User, inverse_of: :lends
  belongs_to :book, required: true

  default_scope { order(created_at: :asc) }
  scope :wait_list, -> { where(borrowed_at: nil) }
  scope :borrowed, -> { where.not(borrowed_at: nil).where(returned_at: nil) }
  scope :returned, -> { where.not(returned_at: nil) }
end

class User < ActiveRecord::Base
  has_many :books, dependent: :destroy
  has_many :shelves, -> { order(position: :asc) }, dependent: :destroy
  has_many :borrows
  has_many :lends

  validates :email, presence: true, uniqueness: true
  validates :slug, presence: true, uniqueness: true
  validates :description, length: { maximum: 500 }
end
