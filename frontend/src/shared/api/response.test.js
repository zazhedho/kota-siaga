import { readData, readPage } from './response'

describe('response readers', () => {
  it('readData returns data when status is true', () => {
    expect(readData({ status: true, data: [1, 2, 3] })).toEqual([1, 2, 3])
  })

  it('readData returns null when status is false or missing', () => {
    expect(readData({ status: false, message: 'err' })).toBeNull()
    expect(readData(null)).toBeNull()
    expect(readData(undefined)).toBeNull()
  })

  it('readPage parses pagination envelope correctly', () => {
    const body = {
      code: 200,
      status: true,
      data: [{ id: 1 }, { id: 2 }],
      total_data: 50,
      total_pages: 5,
      current_page: 2,
      limit: 10,
      next_page: true,
      prev_page: true,
    }

    const page = readPage(body)
    expect(page).toEqual({
      rows: [{ id: 1 }, { id: 2 }],
      total: 50,
      totalPages: 5,
      page: 2,
      limit: 10,
      nextPage: true,
      prevPage: true,
    })
  })

  it('readPage handles empty or null data safely', () => {
    const page = readPage(null)
    expect(page).toEqual({
      rows: [],
      total: 0,
      totalPages: 0,
      page: 1,
      limit: 0,
      nextPage: false,
      prevPage: false,
    })
  })
})
