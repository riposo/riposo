require_relative '../integration_helper'

RSpec.describe 'batch' do
  include IntegerationHelper

  let(:sub_headers) do
    %i[
      Content-Type
      Etag
      Last-Modified
    ]
  end

  it 'POST /v1/batch (empty)' do
    post '/v1/batch', body: {
      requests: [],
    }
    expect(response.status).to eq(200)
    expect(response.headers).to include(
      'Access-Control-Allow-Origin' => '*',
      'Content-Length'              => an_instance_of(String),
      'Content-Security-Policy'     => "default-src 'none'; frame-ancestors 'none'; base-uri 'none';",
      'Content-Type'                => a_string_starting_with('application/json'),
      'Date'                        => an_instance_of(String),
      'X-Content-Type-Options'      => 'nosniff',
    )
    expect(response.data).to match(
      responses: [],
    )
  end

  it 'POST /v1/batch (recursive)' do
    post '/v1/batch', body: {
      requests: [
        { method: 'GET', path: '/' },
        { method: 'POST', path: '/batch' },
      ],
    }
    expect(response.status).to eq(400)
    expect(response.data).to match(
      code: 400,
      details: [{ description: 'Recursive call on /batch endpoint is forbidden.', location: 'body', name: 'requests' }],
      errno: 107,
      error: 'Invalid parameters',
      message: 'requests in body: Recursive call on /batch endpoint is forbidden.',
    )
  end

  it 'POST /v1/batch (40x)' do
    post '/v1/batch', body: {
      requests: [
        { method: 'PUT', path: '/buckets' },
        { method: 'GET', path: '/invalid' },
      ],
    }
    # expect(response.status).to eq(200)
    expect(response.data).to match(
      responses: [
        {
          status: 405,
          path: '/v1/buckets',
          body: {
            code: 405,
            errno: 115,
            error: 'Method Not Allowed',
            message: 'Method not allowed on this endpoint.',
          },
          headers: a_hash_including('Content-Type': instance_of(String)),
        }, {
          status: 404,
          path: '/v1/invalid',
          body: {
            code: 404,
            errno: 111,
            error: 'Not Found',
            message: 'The resource you are looking for could not be found.',
          },
          headers: a_hash_including('Content-Type': instance_of(String)),
        },
      ],
    )
  end

  it 'POST /v1/batch (valid)' do
    post '/v1/batch', body: {
      defaults: {
        method: 'GET',
        path: '/buckets',
      },
      requests: [
        { path: "/v1/buckets/#{bucket}" },
        {},
        { method: 'POST', path: "/buckets/#{bucket}/groups", body: { data: {} } },
      ],
    }
    expect(response.status).to eq(200)
    expect(response.headers).to include(
      'Access-Control-Allow-Origin' => '*',
      'Content-Security-Policy'     => "default-src 'none'; frame-ancestors 'none'; base-uri 'none';",
      'Content-Type'                => a_string_starting_with('application/json'),
      'X-Content-Type-Options'      => 'nosniff',
    )
    expect(response.headers).not_to include('Etag', 'Last-Modified')
    expect(response.data).to match(responses: an_instance_of(Array))
    expect(response.data[:responses]).to match([
      {
        status: 200,
        path: "/v1/buckets/#{bucket}",
        headers: include(:'Cache-Control', *sub_headers),
        body: {
          data: {
            id: bucket,
            last_modified: just_now,
          },
          permissions: {
            write: ["account:#{user}"],
          },
        },
      }, {
        status: 200,
        path: '/v1/buckets',
        headers: include(:'Cache-Control', *sub_headers),
        body: {
          data: [
            include(id: bucket, last_modified: an_instance_of(Integer)),
          ],
        },
      }, {
        status: 201,
        path: "/v1/buckets/#{bucket}/groups",
        headers: include(*sub_headers),
        body: {
          data: {
            id: an_instance_of(String),
            last_modified: just_now,
            members: [],
          },
          permissions: {
            write: ["account:#{user}"],
          },
        },
      },
    ])
  end
end
