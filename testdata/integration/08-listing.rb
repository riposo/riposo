require_relative '../integration_helper'

RSpec.describe 'listing' do
  include IntegerationHelper

  it 'GET /v1/buckets/ID/collections/ID/records' do
    path = "/v1/buckets/#{bucket}/collections/#{random_name}/records"

    # seed records
    post '/v1/batch', body: {
      defaults: {
        method: 'POST',
        path: path.delete_prefix('/v1'),
      },
      requests: [
        { method: 'PUT', path: path.delete_prefix('/v1').delete_suffix('/records'), body: { data: {} } },
        { body: { data: { str: 'o', num: 33, tru: true, nan: 'u' } } },
        { body: { data: { str: 'x', num: 22, tru: false } } },
      ],
    }
    expect(response.status).to eq(200)
    expect(response.data).to match(responses: an_instance_of(Array))
    expect(response.data[:responses].size).to eq(3)

    rec1 = response.data.dig(:responses, 1, :body, :data)
    rec2 = response.data.dig(:responses, 2, :body, :data)

    # paginate records
    get! path, query: { _sort: 'num', _limit: 1 }
    expect(response.headers).to include('Next-Page')
    expect(response.data).to match(data: [rec2])

    get! response.headers['Next-Page'].sub(root_url, '').delete_prefix('/')
    expect(response.headers).not_to include('Next-Page')
    expect(response.data).to match(data: [rec1])

    # filter EQ
    get! path, query: { str: 'x' }
    expect(response.data).to match(data: [rec2])

    # filter NOT
    get! path, query: { not_str: 'x' }
    expect(response.data).to match(data: [rec1])

    # filter HAS
    get! path, query: { has_nan: true }
    expect(response.data).to match(data: [rec1])

    # filter GT
    get! path, query: { gt_str: 'u' }
    expect(response.data).to match(data: [rec2])
    get! path, query: { gt_num: 30 }
    expect(response.data).to match(data: [rec1])
    get! path, query: { gt_nan: 'x' }
    expect(response.data).to match(data: [rec2])
    get! path, query: { gt_nan: 'o' }
    expect(response.data).to match(data: match_array([rec1, rec2]))

    # filter LT
    get! path, query: { lt_str: 'u' }
    expect(response.data).to match(data: [rec1])
    get! path, query: { lt_num: 30 }
    expect(response.data).to match(data: [rec2])
    get! path, query: { lt_nan: 'x' }
    expect(response.data).to match(data: [rec1])
    get! path, query: { lt_nan: 'o' }
    expect(response.data).to match(data: [])
  end
end
