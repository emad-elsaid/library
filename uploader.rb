require 'securerandom'

ALLOWED_IMAGE_TYPES = [:gif, :png, :jpeg]
# orientation = :portrait | nil
def upload_image(uploaded_image, destination, width, height, quality, orientation = :portrait)
  allowed_type = ALLOWED_IMAGE_TYPES.include? FastImage.type(uploaded_image)
  raise ArgumentError, 'File type is not an image. Allowed type JPG, GIF, PNG' unless allowed_type

  size = FastImage.size(uploaded_image)
  raise ArgumentError, 'Cannot get image size' unless size

  w, h = size[0], size[1]
  raise ArgumentError, 'Image should be a portrait. i.e. Width < height' if orientation == :portrait && w > h

  uuid =  SecureRandom.uuid
  image_path = "#{destination}/#{uuid}"

  FileUtils.mv uploaded_image, image_path

  begin
    `mogrify -resize #{width}x#{height}\\> -quality #{quality} -auto-orient -strip #{image_path}`
  rescue
    File.delete(image_path) rescue nil
    raise IOError, "Encountered an error while processing the image"
  end

  uuid
ensure
  File.delete(uploaded_image) rescue nil
end
