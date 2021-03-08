require_relative '../integration_helper'

RSpec.describe 'basics' do
  include IntegerationHelper

  before do
    http.data[:headers].delete('Authorization')
  end

  it 'GET /unknown' do
    get '/unknown'
    expect(response.status).to eq(307)
    expect(response.headers).to include('Location' => "#{root_url}/v1/unknown").or include('Location' => '/v1/unknown')
  end

  it 'GET /v1' do
    get '/v1'
    expect(response.status).to eq(307)
    expect(response.headers).to include('Location' => "#{root_url}/v1/").or include('Location' => '/v1/')
  end

  it 'GET /v1/' do
    get '/v1/'
    expect(response.status).to eq(200)
    expect(response.data).to include(
      project_name: a_string_matching(/\w+/),
      project_version: a_string_matching(/\d+.\d+.\d+/),
      project_docs: a_string_starting_with('https://'),
      http_api_version: '1.22',
      url: a_string_ending_with('/v1/'),
      settings: an_instance_of(Hash),
    )
  end

  it 'POST /v1/' do
    post '/v1/'
    expect(response.status).to eq(405)
    expect(response.headers).to include(
      'Access-Control-Allow-Origin' => '*',
      'Content-Length'              => an_instance_of(String),
      'Content-Security-Policy'     => "default-src 'none'; frame-ancestors 'none'; base-uri 'none';",
      'Content-Type'                => a_string_starting_with('application/json'),
      'Date'                        => an_instance_of(String),
      'X-Content-Type-Options'      => 'nosniff',
    )
    expect(response.data).to include(
      code: 405,
      errno: 115,
      error: 'Method Not Allowed',
      message: 'Method not allowed on this endpoint.',
    )
  end

  it 'GET /v1/unknown' do
    get '/v1/unknown'
    expect(response.status).to eq(404)
    expect(response.data).to include(
      code: 404,
      errno: 111,
      error: 'Not Found',
      message: 'The resource you are looking for could not be found.',
    )
  end

  it 'GET /v1/__heartbeat__' do
    get '/v1/__heartbeat__'
    expect(response.status).to eq(200)
    expect(response.data).to match(
      storage: true,
      permission: true,
      cache: true,
    )
  end

  it 'GET /v1/buckets' do
    get '/v1/buckets'
    expect(response.status).to eq(401)
    expect(response.data).to match(
      code: 401,
      errno: 104,
      error: 'Unauthorized',
      message: 'Please authenticate yourself to use this endpoint.',
    )
  end

  it 'OPTIONS /v1/buckets' do
    options '/v1/buckets'
    expect(response.status).to eq(200)
    expect(response.headers).to include(
      'Access-Control-Allow-Methods' => an_instance_of(String),
      'Access-Control-Allow-Origin'  => an_instance_of(String),
      'Access-Control-Max-Age'       => '3600',
    )
  end

  it 'POST /v1/accounts' do
    post '/v1/accounts', body: {
      data: {},
    }
    expect(response.status).to eq(400)
    expect(response.data).to match(
      code: 400,
      details: [{ description: a_string_matching('Required'), location: 'body', name: 'data.id' }],
      errno: 107,
      error: 'Invalid parameters',
      message: 'data.id in body: Required',
    )

    post '/v1/accounts', body: {
      data: {
        id: user,
      },
    }
    expect(response.status).to eq(400)
    expect(response.data).to match(
      code: 400,
      details: [{ description: 'Required', location: 'body', name: 'data.password' }],
      errno: 107,
      error: 'Invalid parameters',
      message: 'data.password in body: Required',
    )

    post '/v1/accounts', body: {
      data: {
        id: user,
        password: pass,
      },
    }
    expect(response.status).to eq(201)
    expect(response.data).to match(
      permissions: { write: ["account:#{user}"] },
      data: {
        id: user,
        password: an_instance_of(String),
        last_modified: just_now,
      },
    )

    post '/v1/accounts', body: {
      data: {
        id: user,
        password: pass,
      },
    }
    expect(response.status).to eq(401)
    expect(response.data).to match(
      code: 401,
      errno: 104,
      error: 'Unauthorized',
      message: 'Please authenticate yourself to use this endpoint.',
    )
  end

  it 'PUT /v1/accounts/ID' do
    put "/v1/accounts/#{user}", body: {
      data: { password: pass },
    }
    expect(response.status).to eq(401)
  end

  it 'PUT /v1/accounts' do
    put '/v1/accounts', body: {
      data: {
        id: user,
        password: pass,
      },
    }
    expect(response.status).to eq(405)
    expect(response.data).to match(
      code: 405,
      errno: 115,
      error: 'Method Not Allowed',
      message: 'Method not allowed on this endpoint.',
    )
  end
end
