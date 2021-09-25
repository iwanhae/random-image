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
    const [col, setCol] = useState(2)

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
        {[1, 2, 3, 4, 5].map((v) => {
            return <Button key={`b-${v}`} variant="outlined" onClick={() => { setCol(v) }}>{v}</Button>
        })}
        <ImageList variant="masonry" cols={col} gap={8}>
            {imageList.map((item) => {
                return <ImageListItem key={item.id}>
                    <img
                        src={`/data/${item.id}`}
                        loading="eager"
                        className="MuiImageListItem-img"
                    ></img> </ImageListItem>

            })}
        </ImageList>


    </div>
}

export default GroupPage