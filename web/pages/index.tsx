import { Button, ImageList, ImageListItem } from '@mui/material'
import type { NextPage } from 'next'
import Link from 'next/link'
import Head from 'next/head'
import { useEffect, useState } from 'react'
import InfiniteScroll from 'react-infinite-scroller';

type ObjectMeta = {
  id: string;
  name: string;
  group: string
}

let onLoading = 0

const Home: NextPage = () => {
  const [imageList, setImageList] = useState<ObjectMeta[][]>([])
  const [col, setCol] = useState(3)

  const LoadMore = () => {
    console.log("req")
    fetch("/api/sample").then(async (v) => {
      const data: ObjectMeta[] = await v.json()
      setImageList([...imageList, data])
    }).catch((err) => {
      console.log(err)
    }).finally(() => {
    })
  }

  useEffect(LoadMore, [])


  return (
    <div>
      <Head>
        <title>random-image</title>
        <meta name="description" content="showing random image" />
        <link rel="icon" href="/favicon.ico" />
      </Head>

      <main>
        <InfiniteScroll
          initialLoad={false}
          loadMore={LoadMore}
          hasMore={true || false}
          loader={<h1 key={0}>Loading ...</h1>}
          useWindow={true}
        >
          {imageList.map((imgSet, i) => {
            return <div key={`imgset-${i}`}>
              {[1, 2, 3, 4, 5].map((v) => {
                return <Button key={`b-${v}`} variant="outlined" onClick={() => { setCol(v) }}>{v}</Button>
              })}
              <ImageList variant="masonry" cols={col} gap={8} key={i}>
                {imgSet.map((item) => {
                  return <ImageListItem key={item.id}>
                    <Link href={`/g/${item.group}`}>
                      <a href={`/g/${item.group}`}>
                        <img
                          src={`/data/${item.id}`}
                          loading="eager"
                          className="MuiImageListItem-img"
                        ></img>
                      </a>
                    </Link>
                  </ImageListItem>
                })}
              </ImageList>
            </div>
          })}
        </InfiniteScroll>
      </main>
    </div>
  )
}

export default Home
