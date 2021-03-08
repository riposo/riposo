require_relative '../integration_helper'

RSpec.describe 'records' do
  include IntegerationHelper

  it 'POST /v1/buckets/ID/collections/ID/records' do
    # create a record
    post "/v1/buckets/#{bucket}/collections/collection/records", body: {
      data: { extra: 'value' },
    }
    expect(response.status).to eq(201)
    expect(response.headers).to include(
      'Access-Control-Allow-Origin' => '*',
      'Content-Security-Policy'     => "default-src 'none'; frame-ancestors 'none'; base-uri 'none';",
      'Content-Type'                => a_string_starting_with('application/json'),
      'Etag'                        => an_instance_of(String),
      'Last-Modified'               => an_instance_of(String),
      'X-Content-Type-Options'      => 'nosniff',
    )
    expect(response.data).to match(
      permissions: { write: ["account:#{user}"] },
      data: {
        id: an_instance_of(String),
        last_modified: just_now,
        extra: 'value',
      },
    )

    # create a record with explicit ID
    post "/v1/buckets/#{bucket}/collections/collection/records", body: {
      data: {
        id: 'record',
        extra: 'value',
      },
    }
    expect(response.status).to eq(201)
    expect(response.data).to match(
      permissions: { write: ["account:#{user}"] },
      data: {
        id: 'record',
        last_modified: just_now,
        extra: 'value',
      },
    )

    # return record if it already exists
    post "/v1/buckets/#{bucket}/collections/collection/records", body: {
      data: {
        id: 'record',
        extra: 'ignored',
        added: 'ignored',
      },
    }
    expect(response.status).to eq(200)
    expect(response.data).to match(
      permissions: { write: ["account:#{user}"] },
      data: {
        id: 'record',
        last_modified: just_now,
        extra: 'value',
      },
    )

    # no such collection
    post "/v1/buckets/#{bucket}/collections/unknown/records", body: {}
    expect(response.status).to eq(404)
    expect(response.data).to match(
      code: 404,
      details: { id: 'unknown', resource_name: 'collection' },
      errno: 111,
      error: 'Not Found',
    )

    # no access to bucket
    post '/v1/buckets/unknown/collections/unknown/records', body: {}
    expect(response.status).to eq(403).or eq(404) # Kinto BUG: exposing information user should have no knowledge of!
  end

  it 'PUT /v1/buckets/ID/collections/ID/records/ID' do
    # unknown collection
    put "/v1/buckets/#{bucket}/collections/unknown/records/record"
    expect(response.status).to eq(404)
    expect(response.data).to match(
      code: 404,
      details: { id: 'unknown', resource_name: 'collection' },
      errno: 111,
      error: 'Not Found',
    )

    # create new record via PUT
    put "/v1/buckets/#{bucket}/collections/collection/records/other", body: {
      data: { extra: 'something' },
    }
    expect(response.status).to eq(201)
    expect(response.data).to match(
      permissions: { write: ["account:#{user}"] },
      data: {
        id: 'other',
        extra: 'something',
        last_modified: just_now,
      },
    )

    # update existing record
    put "/v1/buckets/#{bucket}/collections/collection/records/record", body: {
      data: {
        extra: 'updated',
        added: 'value',
        last_modified: 1111111111000,
      },
    }
    expect(response.status).to eq(200)
    expect(response.data).to match(
      permissions: { write: ["account:#{user}"] },
      data: {
        id: 'record',
        last_modified: just_now,
        extra: 'updated',
        added: 'value',
      },
    )

    # ID mismatch
    put "/v1/buckets/#{bucket}/collections/collection/records/record", body: {
      data: { id: 'mismatch' },
    }
    expect(response.status).to eq(400)
    expect(response.data).to match(
      code: 400,
      details: a_collection_having(1).items,
      errno: 107,
      error: 'Invalid parameters',
      message: a_string_matching(/not match .+ object/),
    )

    # invalid collection
    put "/v1/buckets/#{bucket}/collections/unknown/records/record", body: {
      data: {},
    }
    expect(response.status).to eq(404)
    expect(response.data).to match(
      code: 404,
      details: { id: 'unknown', resource_name: 'collection' },
      errno: 111,
      error: 'Not Found',
    )

    # no access to bucket
    put '/v1/buckets/unknown/collections/unknown/records/record'
    expect(response.status).to eq(403).or eq(404) # Kinto BUG: exposing information user should have no knowledge of!
  end

  it 'PATCH /v1/buckets/ID/collections/ID/records/ID' do
    # unknown record
    patch "/v1/buckets/#{bucket}/collections/collection/records/unknown", body: {
      data: {},
    }
    expect(response.status).to eq(404)
    expect(response.data).to match(
      code: 404,
      details: { id: 'unknown', resource_name: 'record' },
      errno: 110,
      error: 'Not Found',
    )

    # unknown collection
    patch "/v1/buckets/#{bucket}/collections/unknown/records/record", body: {
      data: {},
    }
    expect(response.status).to eq(404)
    expect(response.data).to match(
      code: 404,
      details: { id: 'unknown', resource_name: 'collection' },
      errno: 111,
      error: 'Not Found',
    )

    # unknown record in unknown collection
    patch "/v1/buckets/#{bucket}/collections/unknown/records/unknown", body: {
      data: {},
    }
    expect(response.status).to eq(404)
    expect(response.data).to match(
      code: 404,
      details: { id: 'unknown', resource_name: 'collection' },
      errno: 111,
      error: 'Not Found',
    )

    # patch record
    patch "/v1/buckets/#{bucket}/collections/collection/records/record", body: {
      data: {
        last_modified: 1111111111000,
        added: 'updated',
      },
    }
    expect(response.status).to eq(200)
    expect(response.data).to match(
      permissions: { write: ["account:#{user}"] },
      data: {
        id: 'record',
        last_modified: just_now,
        extra: 'updated',
        added: 'updated',
      },
    )

    # ID mismatch
    patch "/v1/buckets/#{bucket}/collections/collection/records/record", body: {
      data: { id: 'mismatch' },
    }
    expect(response.status).to eq(400)
    expect(response.data).to match(
      code: 400,
      details: a_collection_having(1).items,
      errno: 107,
      error: 'Invalid parameters',
      message: a_string_matching(/not match .+ object/),
    )

    # invalid collection
    patch "/v1/buckets/#{bucket}/collections/unknown/records/record", body: {
      data: {},
    }
    expect(response.status).to eq(404)
    expect(response.data).to match(
      code: 404,
      details: { id: 'unknown', resource_name: 'collection' },
      errno: 111,
      error: 'Not Found',
    )

    # no access to bucket
    patch '/v1/buckets/unknown/collections/unknown/records/record', body: {
      data: {},
    }
    expect(response.status).to eq(403).or eq(404) # Kinto BUG: exposing information user should have no knowledge of!
  end

  it 'DELETE /v1/buckets/ID/collections/ID/records/ID' do
    # delete a single record
    delete "/v1/buckets/#{bucket}/collections/collection/records/other"
    expect(response.status).to eq(200)
    expect(response.data).to match(
      data: {
        deleted: true,
        id: 'other',
        last_modified: just_now,
      },
    )

    # fail if record is missing
    delete "/v1/buckets/#{bucket}/collections/collection/records/other"
    expect(response.status).to eq(404)
    expect(response.data).to match(
      code: 404,
      details: { id: 'other', resource_name: 'record' },
      errno: 110,
      error: 'Not Found',
    )

    # invalid collection
    delete "/v1/buckets/#{bucket}/collections/unknown/records/record"
    expect(response.status).to eq(404)
    expect(response.data).to match(
      code: 404,
      details: { id: 'unknown', resource_name: 'collection' },
      errno: 111,
      error: 'Not Found',
    )

    # no access to bucket
    delete '/v1/buckets/unknown/collections/unknown/records/record'
    expect(response.status).to eq(403).or eq(404) # Kinto BUG: exposing information user should have no knowledge of!
  end

  it 'GET /v1/buckets/ID/collections/ID/records' do
    # list records
    get "/v1/buckets/#{bucket}/collections/collection/records"
    expect(response.status).to eq(200)
    expect(response.data).to match(
      data: a_collection_having(2).items,
    )

    # invalid collection
    get "/v1/buckets/#{bucket}/collections/unknown/records"
    expect(response.status).to eq(404)
    expect(response.data).to match(
      code: 404,
      details: { id: 'unknown', resource_name: 'collection' },
      errno: 111,
      error: 'Not Found',
    )

    # no access to bucket
    get '/v1/buckets/unknown/collections/unknown/records'
    expect(response.status).to eq(403)
  end

  it 'GET /v1/buckets/ID/collections/ID/records/ID' do
    # fetch a single record
    get "/v1/buckets/#{bucket}/collections/collection/records/record"
    expect(response.status).to eq(200)
    expect(response.data).to match(
      permissions: { write: ["account:#{user}"] },
      data: {
        id: 'record',
        last_modified: just_now,
        extra: 'updated',
        added: 'updated',
      },
    )

    # record is missing
    get "/v1/buckets/#{bucket}/collections/collection/records/unknown"
    expect(response.status).to eq(404)
    expect(response.data).to match(
      code: 404,
      details: { id: 'unknown', resource_name: 'record' },
      errno: 110,
      error: 'Not Found',
    )

    # invalid collection
    get "/v1/buckets/#{bucket}/collections/unknown/records/record"
    expect(response.status).to eq(404)
    expect(response.data).to match(
      code: 404,
      details: { id: 'unknown', resource_name: 'collection' },
      errno: 111,
      error: 'Not Found',
    )

    # no access to bucket
    get '/v1/buckets/unknown/collections/unknown/records/record'
    expect(response.status).to eq(403)
  end
end
