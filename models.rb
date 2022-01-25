require 'securerandom'
require_relative 'uploader'

class Book < ActiveRecord::Base
  IMAGES_PATH = 'public/books/image'

  belongs_to :shelf
  belongs_to :user, required: true
  has_many :highlights, dependent: :destroy

  validates_presence_of :title, :author, :isbn
  validates :google_books_id, length: { in: 0..30, allow_nil: true }
  validates :isbn, length: { is: 13 }, numericality: true, uniqueness: { scope: :user }
  validate :isbn13_format

  default_scope { order(created_at: :desc) }
  before_destroy { File.delete image_path if image? && File.exist?(image) }

  def isbn13_format
    digits = isbn.chars.map(&:to_i)
    rem = digits.map.with_index { |digit, index| index.even? ? digit : digit * 3 }.reduce(:+) % 10
    errors.add(:isbn, 'value is not a valid ISBN13') unless rem.zero?
  end

  def image_path
    "#{IMAGES_PATH}/#{image}"
  end

  def upload(uploaded_image)
    name = upload_image(uploaded_image, IMAGES_PATH, 432, 576, 60, :portrait)
    File.delete(image_path) rescue nil
    update(image: name)
  rescue StandardError => e
    errors.add(:image, e.message)
    false
  end
end

class Shelf < ActiveRecord::Base
  belongs_to :user, required: true
  has_many :books, dependent: :nullify
  acts_as_list scope: :user

  validates_presence_of :name
end

class User < ActiveRecord::Base
  PROFILES = [:facebook, :twitter, :linkedin, :instagram, :phone, :whatsapp, :telegram]
  has_many :books, dependent: :destroy
  has_many :shelves, -> { order(position: :asc) }, dependent: :destroy
  has_many :emails, dependent: :destroy

  validates :email, presence: true, uniqueness: true
  validates :slug, presence: true, uniqueness: true
  validates :description, length: { maximum: 500 }
  validates :amazon_associates_id, length: { maximum: 50 }

  def self.signup(name, email, image)
    attrs = { name: name, image: image, slug: SecureRandom.uuid }
    User.create_with(attrs).find_or_create_by(email: email)
  end
end

class Email < ActiveRecord::Base
  belongs_to :user, required: true
  belongs_to :emailable, polymorphic: true, required: true

  validates :about, presence: true
end

class Highlight < ActiveRecord::Base
  IMAGES_SERVE_PATH = '/highlights/image'
  IMAGES_PATH = "public#{IMAGES_SERVE_PATH}"

  belongs_to :book, required: true

  before_validation :reformat_content

  validates :content, length: { minimum: 20, maximum: 2000 }
  validates :page, presence: true

  default_scope { order(:page, :created_at) }

  def image_path
    "#{IMAGES_PATH}/#{image}"
  end

  def image_url
    "#{IMAGES_SERVE_PATH}/#{image}"
  end

  def upload(uploaded_image)
    name = upload_image(uploaded_image, IMAGES_PATH, 600, 600, 60, :nil)
    File.delete(image_path) rescue nil
    update(image: name)
  rescue StandardError => e
    errors.add(:image, e.message)
    false
  end

  # TODO review any multiline input on the other models and have this function for it. and use simple format to print them
  def reformat_content
    self.content = content.to_s.lines.map(&:strip).join("\n").gsub(/\n{3,}/, "\n\n")
  end
end
