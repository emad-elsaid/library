# frozen_string_literal: true

require 'rack/utils'

helpers do
  def partial (template, locals = {})
    erb(template.to_sym, layout: false, locals: locals)
  end

  def h(text)
    Rack::Utils.escape_html(text)
  end

  def guest?
    current_user.nil?
  end

  def loggedin?
    !current_user.nil?
  end

  def book_cover(book)
    return "/books/image/#{book.image}" if book.image?
    return "https://books.google.com/books/content?id=#{book.google_books_id}&printsec=frontcover&img=1&zoom=1" if book.google_books_id?

    '/default_book'
  end

  def current_user
    return unless session[:user]

    @current_user ||= User.find(session[:user])
  end

  def can?(verb, record = nil)
    case record

    when nil
      case verb
      when :login then guest?
      when :logout then loggedin?
      when :list_access then loggedin?
      when :to_lend then loggedin?
      when :to_borrow then loggedin?
      else raise "Verb #{verb} not handled for #{record}"
      end

    when Book
      case verb
      when :create then record.user_id == current_user&.id
      when :edit then record.user_id == current_user&.id
      when :delete then record.user_id == current_user&.id
      when :borrow then
        loggedin? &&
          record.user_id != current_user&.id &&
          !record.borrows.exists?(user: current_user) &&
          current_user.accesses_from.accepted.exists?(owner: record.user_id)
      else raise "Verb #{verb} not handled for #{record}"
      end

    when Shelf
      case verb
      when :create then record.user_id == current_user&.id
      when :edit then record.user_id == current_user&.id
      when :delete then record.user_id == current_user&.id
      else raise "Verb #{verb} not handled for #{record}"
      end

    when User
      case verb
      when :edit then record == current_user
      when :access then loggedin? && (record == current_user || current_user.accesses_from.exists?(owner: record))
      when :list_shelves then loggedin? && record == current_user
      else raise "Verb #{verb} not handled for #{record}"
      end

    when Borrow
      case verb
      when :show then loggedin?
      when :delete then (record.user_id == current_user&.id && record.borrowed_at.nil?) || record.owner_id == current_user&.id
      when :borrow then !record.book.borrows.borrowed.exists? && record.owner_id == current_user&.id
      when :return then record.borrowed_at.present? && record.owner_id == current_user&.id
      when :contact then record.owner_id == current_user&.id
      else raise "Verb #{verb} not handled for #{record}"
      end

    when Access
      case verb
      when :show then record.user_id == current_user&.id || record.owner_id == current_user&.id
      when :create then record.user_id == current_user&.id && !Access.exists?(user: record.user_id, owner: record.owner_id)
      when :accept then record.owner_id == current_user&.id && record.accepted_at.nil?
      when :reject then record.owner_id == current_user&.id && record.rejected_at.nil?
      when :delete then record.user_id == current_user&.id || record.owner_id == current_user&.id
      when :contact then record.owner_id == current_user&.id
      else raise "Verb #{verb} not handled for #{record}"
      end

    else raise "Error #{record} permissions not handled."
    end
  end

  def format_date(date)
    date.to_date
  end

  def meta_property(name)
    return unless @meta
    return unless @meta.key?(name)

    "<meta property=\"#{name}\" value=\"#{h @meta[name]}\"/>"
  end

  def meta_name(name)
    return unless @meta
    return unless @meta.key?(name)

    "<meta name=\"#{name}\" value=\"#{h @meta[name]}\"/>"
  end

  def set_meta(name, value)
    @meta ||= {}
    @meta[name] = value
  end
end
