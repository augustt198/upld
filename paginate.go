package main

import (
    "strconv"
    
    "gopkg.in/mgo.v2"
)

const rangeSep string =
    `<a class="btn btn-outline disabled">...</a>`

func Paginate(page int, perPage int, q *mgo.Query) *mgo.Query {
    return q.Limit(perPage).Skip((page - 1) * perPage)
}

func PaginateBar(page int, count int) string {
    html := `<div class="btn-group">
        <a class="btn btn-outline`

    if page < 2 {
        html += ` disabled">Previous</a>`        
    } else {
        html += `" href="?page=` + strconv.Itoa(page - 1) + `">
            Previous</a>`
    }


    if count < 10 {
        html += btnRange(1, count, page)
    } else {
        if page < 7 {
            html += btnRange(1, 8, page)
            html += rangeSep
            html += btnRange(count - 1, count, page)
        } else if (page > count - 6) {
            html += btnRange(1, 2, page)
            html += rangeSep
            html += btnRange(page - 2, count, page)
        } else {
            html += btnRange(1, 2, page)
            html += rangeSep
            html += btnRange(page - 2, page + 2, page)
            html += rangeSep
            html += btnRange(count - 1, count, page)
        }
    }

    html += `<a class="btn btn-outline`
    if page >= count {
        html += ` disabled">Next</a>`
    } else {
        html += `" href="?page=` + strconv.Itoa(page + 1) + `">
            Next</a>`
    }

    return html
}

func btnRange(start int, limit int, selected int) string {
    html := ""
    for i := start; i <= limit; i++ {
        html += btnFor(i, selected)
    }

    return html
}

func btnFor(page int, selected int) string {
    html := `<a href="?page=` + strconv.Itoa(page) + `" class="btn btn-outline`
    if page == selected {
        html += ` selected">` + strconv.Itoa(page) + `</a>`
    } else {
        html += `">` + strconv.Itoa(page) + `</a>`
    }

    return html
}
