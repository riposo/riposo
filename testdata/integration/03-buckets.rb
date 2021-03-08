require_relative '../integration_helper'

RSpec.describe 'buckets' do
  include IntegerationHelper

  it 'GET /v1/buckets' do
    get '/v1/buckets'
    expect(response.status).to eq(200)
    expect(response.headers).to include(
      'Access-Control-Allow-Origin' => '*',
      'Cache-Control'               => a_string_starting_with('no-cache, no-store'),
      'Content-Security-Policy'     => "default-src 'none'; frame-ancestors 'none'; base-uri 'none';",
      'Content-Type'                => a_string_starting_with('application/json'),
      'Etag'                        => an_instance_of(String),
      'Last-Modified'               => an_instance_of(String),
      'X-Content-Type-Options'      => 'nosniff',
    )
    expect(response.data).to match(
      data: [],
    )
  end

  it 'POST /v1/buckets' do
    # create bucket with specific ID
    post '/v1/buckets', body: {
      data: { id: bucket },
    }
    expect(response.status).to eq(201)
    expect(response.data).to match(
      permissions: { write: ["account:#{user}"] },
      data: {
        id: bucket,
        last_modified: just_now,
      },
    )

    # check result
    get '/v1/buckets'
    expect(response.status).to eq(200)
    expect(response.data).to match(
      data: [
        {
          id: bucket,
          last_modified: just_now,
        },
      ],
    )

    # try again with etag
    get '/v1/buckets', headers: { 'If-None-Match' => response.etag }
    expect(response.status).to eq(304)

    get '/v1/buckets', headers: { 'If-Match' => '"1111111111000"' }
    expect(response.status).to eq(412)
    expect(response.data).to match(
      code: 412,
      errno: 114,
      error: 'Precondition Failed',
      message: 'Resource was modified meanwhile',
    )
  end

  it 'HEAD /v1/buckets' do
    head '/v1/buckets'
    expect(response.status).to eq(200)
    expect(response.headers).to include(
      'Access-Control-Allow-Origin' => '*',
      'Cache-Control'               => a_string_starting_with('no-cache, no-store'),
      'Content-Security-Policy'     => "default-src 'none'; frame-ancestors 'none'; base-uri 'none';",
      'Etag'                        => an_instance_of(String),
      'Last-Modified'               => an_instance_of(String),
      'X-Content-Type-Options'      => 'nosniff',
      'Total-Objects'               => '1',
      'Total-Records'               => '1',
    )
    expect(response.body).to be_empty
  end

  it 'GET /v1/buckets/ID' do
    get "/v1/buckets/#{bucket}"
    expect(response.status).to eq(200)
    expect(response.headers).to include(
      'Access-Control-Allow-Origin' => '*',
      'Cache-Control'               => a_string_starting_with('no-cache, no-store'),
      'Content-Security-Policy'     => "default-src 'none'; frame-ancestors 'none'; base-uri 'none';",
      'Content-Type'                => a_string_starting_with('application/json'),
      'Etag'                        => an_instance_of(String),
      'Last-Modified'               => an_instance_of(String),
      'X-Content-Type-Options'      => 'nosniff',
    )
    expect(response.data).to match(
      permissions: { write: ["account:#{user}"] },
      data: {
        id: bucket,
        last_modified: just_now,
      },
    )

    get '/v1/buckets/unknown'
    expect(response.status).to eq(403)
    expect(response.data).to match(
      code: 403,
      errno: 121,
      error: 'Forbidden',
      message: 'This user cannot access this resource.',
    )
  end
end
