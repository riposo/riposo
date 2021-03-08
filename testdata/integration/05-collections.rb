require_relative '../integration_helper'

RSpec.describe 'collections' do
  include IntegerationHelper

  it 'POST /v1/buckets/ID/collections' do
    # reject invalid IDs
    post "/v1/buckets/#{bucket}/collections", body: {
      data: {
        id: 'invalid id',
      },
    }
    expect(response.status).to eq(400)
    expect(response.data).to match(
      code: 400,
      details: [a_hash_including(description: 'Invalid object id', location: 'path')],
      errno: 107,
      error: 'Invalid parameters',
      message: 'path: Invalid object id',
    )

    # create new collections
    post "/v1/buckets/#{bucket}/collections", body: {
      data: {},
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
      },
    )

    # create new collections with explicit ID
    post "/v1/buckets/#{bucket}/collections", body: {
      data: {
        id: 'collection',
        extra: 'value',
      },
    }
    expect(response.status).to eq(201)
    expect(response.data).to match(
      permissions: { write: ["account:#{user}"] },
      data: {
        id: 'collection',
        last_modified: just_now,
        extra: 'value',
      },
    )

    # return existing
    post "/v1/buckets/#{bucket}/collections", body: {
      data: {
        id: 'collection',
        extra: 'ignored',
      },
    }
    expect(response.status).to eq(200)
    expect(response.data).to match(
      permissions: { write: ["account:#{user}"] },
      data: {
        id: 'collection',
        last_modified: just_now,
        extra: 'value',
      },
    )

    # a body is not even required
    post "/v1/buckets/#{bucket}/collections"
    expect(response.status).to eq(201)

    # no access to bucket
    post '/v1/buckets/unknown/collections'
    expect(response.status).to eq(403)
  end

  it 'PUT /v1/buckets/ID/collections/ID' do
    # create collection via PUT
    put "/v1/buckets/#{bucket}/collections/other", body: {
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

    # update collection
    put "/v1/buckets/#{bucket}/collections/collection", body: {
      data: { extra: 'updated' },
      permissions: { 'write' => ['group:peers'], 'read' => ['system.Authenticated'] },
    }
    expect(response.status).to eq(200)
    expect(response.data).to match(
      permissions: {
        read: ['system.Authenticated'],
        write: match_array(['group:peers', "account:#{user}"]),
      },
      data: {
        id: 'collection',
        extra: 'updated',
        last_modified: just_now,
      },
    )

    # no access to bucket
    put '/v1/buckets/unknown/collections/collection'
    expect(response.status).to eq(403)
  end

  it 'PATCH /v1/buckets/ID/collections/ID' do
    patch "/v1/buckets/#{bucket}/collections/collection", body: {
      data: { added: 'value' },
      permissions: { write: [] },
    }
    expect(response.status).to eq(200)
    expect(response.data).to match(
      permissions: {
        read: ['system.Authenticated'],
        write: ["account:#{user}"],
      },
      data: {
        id: 'collection',
        extra: 'updated',
        added: 'value',
        last_modified: just_now,
      },
    )

    # no access to bucket
    patch '/v1/buckets/unknown/collections/collection', body: {
      data: {},
    }
    expect(response.status).to eq(403)
  end

  it 'DELETE /v1/buckets/ID/collections/ID' do
    # delete a collection
    delete "/v1/buckets/#{bucket}/collections/other"
    expect(response.status).to eq(200)
    expect(response.data).to match(
      data: {
        deleted: true,
        id: 'other',
        last_modified: just_now,
      },
    )

    # collection doesn't exist
    delete "/v1/buckets/#{bucket}/collections/other"
    expect(response.status).to eq(404)
    expect(response.data).to match(
      code: 404,
      details: { id: 'other', resource_name: 'collection' },
      errno: 110,
      error: 'Not Found',
    )

    # no access to bucket
    delete '/v1/buckets/unknown/collections/collection'
    expect(response.status).to eq(403)
  end

  it 'GET /v1/buckets/ID/collections' do
    # list collections
    get "/v1/buckets/#{bucket}/collections"
    expect(response.status).to eq(200)
    expect(response.data).to match(
      data: a_collection_having(3).items,
    )

    # no access to bucket
    get '/v1/buckets/unknown/collections'
    expect(response.status).to eq(403)
  end

  it 'GET /v1/buckets/ID/collections/ID' do
    # get a single collection
    get "/v1/buckets/#{bucket}/collections/collection"
    expect(response.status).to eq(200)
    expect(response.data).to match(
      permissions: include(:read, :write),
      data: include(:id, :extra, :last_modified),
    )

    # collection doesn't exist
    get "/v1/buckets/#{bucket}/collections/unknown"
    expect(response.status).to eq(404)
    expect(response.data).to match(
      code: 404,
      details: { id: 'unknown', resource_name: 'collection' },
      errno: 110,
      error: 'Not Found',
    )

    # no access to bucket
    get '/v1/buckets/unknown/collections/unknown'
    expect(response.status).to eq(403)
  end
end
