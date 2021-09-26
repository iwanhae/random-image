import { Button, ImageList, ImageListItem } from "@mui/material";
import { NextPage } from "next";
import Link from 'next/link'
import { useRouter } from "next/dist/client/router";
import { useEffect, useState } from "react";

type ObjectMeta = {
    id: string;
    name: string;
    group: string
}

const GroupPage: NextPage = () => {
    const router = useRouter()
    const { id } = router.query

    const [imageList, setImageList] = useState<ObjectMeta[]>([])
    const [col, setCol] = useState(1)
    const [width, setWidth] = useState(1024)

    useEffect(() => {
        const width = JSON.parse(localStorage.getItem("detail-width") || "1024") as number
        const col = JSON.parse(localStorage.getItem("detail-col") || "1") as number
        setWidth(width)
        setCol(col)
    }, [])

    useEffect(() => {
        localStorage.setItem("detail-width", `${width}`)
        localStorage.setItem("detail-col", `${col}`)
    }, [width, col])

    useEffect(() => {
        const last = window.location.href.split('/').pop();
        fetch(`/api/group/${last}`).then(async (v) => {
            const data: ObjectMeta[] = await v.json()
            setImageList(data)
        }).catch((err) => {
            console.log(err)
        }).finally(() => {

        })
    }, [])
    return <div>
        <h1 style={{ wordWrap: "break-word" }}>{id}</h1>
        <Link href={"/"}><Button variant="contained">Back</Button></Link>
        <div>
            {[1, 2, 3, 4, 5].map((v) => {
                return <Button key={`b-${v}`} variant={v == col ? "contained" : "outlined"} onClick={() => { setCol(v) }}>{v}</Button>
            })}
        </div>
        <div>
            {[480, 720, 1024, 1920, 0].map((v) => {
                return <Button key={`b-${v}`} variant={v == width ? "contained" : "outlined"} onClick={() => { setWidth(v) }}>{v}</Button>
            })}
        </div>
        <ImageList variant="masonry" cols={col} gap={8}>
            {imageList.map((item) => {
                return <ImageListItem key={item.id}>
                    <a href={`/data/${item.id}`} target="_blank" rel="noopener noreferrer">
                        <img
                            src={width == 0 ? `/data/${item.id}` : `/data/${item.id}?q=70&w=${width}`}
                            loading="eager"
                            className="MuiImageListItem-img"
                        ></img>
                    </a></ImageListItem>

            })}
        </ImageList>


    </div>
}

export default GroupPage