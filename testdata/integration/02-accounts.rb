require_relative '../integration_helper'

RSpec.describe 'accounts' do
  include IntegerationHelper

  it 'GET /v1/accounts' do
    get '/v1/accounts'
    expect(response.status).to eq(200)
    expect(response.data).to match(
      data: [
        {
          id: user,
          password: an_instance_of(String),
          last_modified: just_now,
        },
      ],
    )
  end

  it 'GET /v1/accounts/ID' do
    get "/v1/accounts/#{user}"
    expect(response.status).to eq(200)
    expect(response.data).to match(
      permissions: { write: ["account:#{user}"] },
      data: {
        id: user,
        password: an_instance_of(String),
        last_modified: just_now,
      },
    )
  end

  it 'POST /v1/accounts' do
    post '/v1/accounts', body: {
      data: {
        id: user,
        password: pass,
      },
    }
    expect(response.status).to eq(200)
    expect(response.data).to match(
      permissions: { write: ["account:#{user}"] },
      data: {
        id: user,
        password: an_instance_of(String),
        last_modified: just_now,
      },
    )
  end

  it 'PUT /v1/accounts/ID' do
    put "/v1/accounts/#{user}", body: {
      data: { password: pass },
    }
    expect(response.status).to eq(200)
    expect(response.data).to match(
      permissions: { write: ["account:#{user}"] },
      data: {
        id: user,
        password: an_instance_of(String),
        last_modified: just_now,
      },
    )
  end
end
