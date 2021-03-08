require_relative '../integration_helper'

RSpec.describe 'groups' do
  include IntegerationHelper

  it 'POST /v1/buckets/ID/groups' do
    # empty groups are OK
    post "/v1/buckets/#{bucket}/groups", body: {
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
        members: [],
      },
    )

    # group with a member and explicit ID
    post "/v1/buckets/#{bucket}/groups", body: {
      data: {
        id: 'group',
        members: ['account:alice'],
      },
    }
    expect(response.status).to eq(201)
    expect(response.data).to match(
      permissions: { write: ["account:#{user}"] },
      data: {
        id: 'group',
        last_modified: just_now,
        members: ['account:alice'],
      },
    )

    # existing groups are not overwritten
    post "/v1/buckets/#{bucket}/groups", body: {
      data: {
        id: 'group',
        members: ['ignored'],
      },
    }
    expect(response.status).to eq(200)
    expect(response.data).to match(
      permissions: { write: ["account:#{user}"] },
      data: {
        id: 'group',
        last_modified: just_now,
        members: ['account:alice'],
      },
    )

    # no access to bucket
    post '/v1/buckets/unknown/groups', body: {
      data: { members: [] },
    }
    expect(response.status).to eq(403)
  end

  it 'PUT /v1/buckets/ID/groups' do
    # members are not required
    put "/v1/buckets/#{bucket}/groups/group", body: {
      data: {},
    }
    expect(response.status).to eq(200)
    expect(response.data).to match(
      permissions: { write: ["account:#{user}"] },
      data: {
        id: 'group',
        last_modified: just_now,
        members: [],
      },
    )
  end

  it 'PATCH /v1/buckets/ID/groups' do
    # reinstate alice as a member
    patch "/v1/buckets/#{bucket}/groups/group", body: {
      data: { members: ['account:alice'] },
    }
    expect(response.status).to eq(200)
    expect(response.data).to match(
      permissions: { write: ["account:#{user}"] },
      data: {
        id: 'group',
        last_modified: just_now,
        members: ['account:alice'],
      },
    )

    # skip updates if members empty
    patch "/v1/buckets/#{bucket}/groups/group", body: {
      data: {},
    }
    expect(response.status).to eq(200)
    expect(response.data).to match(
      permissions: { write: ["account:#{user}"] },
      data: {
        id: 'group',
        last_modified: just_now,
        members: ['account:alice'],
      },
    )
  end
end
