require 'securerandom'

class Book < ActiveRecord::Base
  ALLOWED_TYPES = [:gif, :png, :jpeg]

  belongs_to :shelf
  belongs_to :user, required: true
  has_many :borrows

  validates_presence_of :title, :author, :isbn

  default_scope { order(created_at: :desc) }
  before_destroy { File.delete image_path if image }

  def image_path
    "public/books/image/#{image}"
  end

  def upload(uploaded_image)
    size = FastImage.size(uploaded_image)
    allowed_type = ALLOWED_TYPES.include? FastImage.type(uploaded_image)

    errors.add(:image, 'File type is not an image. Allowed type JPG, GIF, PNG') unless allowed_type
    errors.add(:image, 'Image should be a portrait. width should be less than height') if size && size[0] > size[1]
    return if invalid?

    File.delete(image_path) if image? && File.exist?(image_path)
    update(image: SecureRandom.uuid)
    FileUtils.mv uploaded_image, image_path

    begin
      `mogrify -resize 432x576\\> -quality 60 -auto-orient -strip #{image_path}`
    rescue e
      puts "Encountered error while processing image for #{id} image: #{image_path}"
    end
  end
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

  scope :by_creation, -> { order(created_at: :asc) }
  scope :wait_list, -> { where(borrowed_at: nil) }
  scope :borrowed, -> { where.not(borrowed_at: nil).where(returned_at: nil) }
  scope :returned, -> { where.not(returned_at: nil) }
end

class Access < ActiveRecord::Base
  belongs_to :user, foreign_key: :user_id, required: true, class_name: :User, inverse_of: :accesses_from
  belongs_to :owner, foreign_key: :owner_id, required: true, class_name: :User, inverse_of: :accesses_to

  default_scope { order(created_at: :desc) }
  scope :pending, -> { where(accepted_at: nil, rejected_at: nil) }
  scope :accepted, -> { where.not(accepted_at: nil) }
  scope :rejected_at, -> { where.not(rejected_at: nil) }
end

class User < ActiveRecord::Base
  PROFILES = [:facebook, :twitter, :linkedin, :instagram, :phone, :whatsapp, :telegram]
  has_many :books, dependent: :destroy
  has_many :shelves, -> { order(position: :asc) }, dependent: :destroy
  has_many :borrows, dependent: :destroy

  has_many :borrow_wait_lists, -> { wait_list }, class_name: :Borrow, foreign_key: :user_id
  has_many :books_to_borrow, -> { distinct }, through: :borrow_wait_lists, source: :book

  has_many :accepted_borrows, -> { borrowed }, class_name: :Borrow, foreign_key: :user_id
  has_many :borrowed_books, through: :accepted_borrows, source: :book

  has_many :lends, dependent: :destroy
  has_many :lend_wait_lists, -> { wait_list }, class_name: :Borrow, foreign_key: :owner_id
  has_many :books_to_lend, -> { distinct }, through: :lend_wait_lists, source: :book

  has_many :accepted_lends, -> { borrowed }, class_name: :Borrow, foreign_key: :owner_id
  has_many :lent_books, through: :accepted_lends, source: :book

  has_many :accesses_from, foreign_key: :user_id, dependent: :destroy, class_name: :Access, inverse_of: :user
  has_many :accesses_to, foreign_key: :owner_id, dependent: :destroy, class_name: :Access, inverse_of: :owner

  validates :email, presence: true, uniqueness: true
  validates :slug, presence: true, uniqueness: true
  validates :description, length: { maximum: 500 }

  def self.signup(name, email, image)
    attrs = { name: name, image: image, slug: SecureRandom.uuid }
    User.create_with(attrs).find_or_create_by(email: email)
  end
end
