export function readData(body) {
  return body?.status === true ? body.data : null
}

export function readPage(body) {
  return {
    rows: Array.isArray(body?.data) ? body.data : [],
    total: body?.total_data ?? 0,
    totalPages: body?.total_pages ?? 0,
    page: body?.current_page ?? 1,
    limit: body?.limit ?? 0,
    nextPage: Boolean(body?.next_page),
    prevPage: Boolean(body?.prev_page),
  }
}
