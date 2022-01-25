user1 = User.create!(
  name: 'emad',
  email: 'emad@example.com',
  slug: SecureRandom.uuid,
  image: 'https://www.emadelsaid.com/images/avatar.webp'
)
user2 = User.create!(
  name: 'github emad',
  email: 'github@example.com',
  slug: SecureRandom.uuid,
  image: 'https://avatars.githubusercontent.com/u/54403?s=60&v=4'
)
user3 = User.create!(
  name: 'random person',
  email: 'random@example.com',
  slug: SecureRandom.uuid
)
self_dev = Shelf.create!(user: user1, name: 'Self Development')
novels = Shelf.create!(user: user1, name: 'Novels')

the_subtle = Book.create!(
  user: user1,
  shelf: self_dev,
  title: 'The subtle art of not giving a fuck',
  author: 'mark manson',
  isbn: 9780062457714
)
