require 'rspec'
require 'rspec/collection_matchers'
require 'excon'
require 'json'
require 'yaml'
require 'securerandom'
require 'base64'

RSpec.configure do |c|
  c.fail_fast = 1
end

module IntegerationHelper
  class Response
    attr_reader :status, :headers, :body

    def initialize(resp)
      @status = resp.status
      @headers = resp.headers
      @body = resp.body
    end

    def data
      @data ||= symbolize(JSON.parse(body))
    end

    def etag
      headers['Etag']
    end

    def last_modified
      headers['Last-Modified']
    end

    private

    def symbolize(object)
      case object
      when Hash
        object.keys.each do |key|
          key = 'Etag' if key == 'ETag' # normalize Etag headers in batch responses
          object[key.to_sym] = symbolize(object.delete(key))
        end
      when Array
        object.map! {|e| symbolize(e) }
      end
      object
    end
  end

  extend RSpec::Matchers::DSL

  ROOT_URL = ENV.fetch('RIPOSO_URL', 'http://localhost:8888').freeze
  USER     = "test-#{SecureRandom.hex(3)}".freeze
  BUCKET   = "#{USER}-bucket".freeze
  PASS     = 's3c6et'.freeze

  def root_url
    ROOT_URL
  end

  def user
    USER
  end

  def pass
    PASS
  end

  def bucket
    BUCKET
  end

  def random_name
    SecureRandom.hex(3)
  end

  def response
    @response
  end

  def http
    @http ||= Excon.new root_url, headers: {
      'Authorization' => "Basic #{Base64.strict_encode64("#{user}:#{pass}")}",
      'Origin'        => root_url,
    }
  end

  def request(method, **opts)
    opts[:headers] ||= {}
    opts[:headers].update('Access-Control-Request-Method' => method.to_s.upcase)
    @response = Response.new(http.request(method: method, **opts))
  end

  def get(path, **opts)
    request :get, path: path, **opts
  end

  def get!(path, **opts)
    get(path, **opts)
    expect(response.status).to eq(200)
  end

  def head(path, **opts)
    request :head, path: path, **opts
  end

  def options(path, **opts)
    request :options, path: path, **opts
  end

  def delete(path, **opts)
    request :delete, path: path, **opts
  end

  def post(path, body: nil, **opts)
    request :post, path: path, body: body&.to_json, **opts
  end

  def put(path, body: nil, **opts)
    request :put, path: path, body: body&.to_json, **opts
  end

  def patch(path, body: nil, **opts)
    request :patch, path: path, body: body&.to_json, **opts
  end

  def just_now
    be_within(3 * 1000).of(Time.now.to_i * 1000)
  end
end
