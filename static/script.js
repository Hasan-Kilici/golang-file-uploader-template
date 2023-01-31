function getImage(id){
let image;
  fetch(`/images/${id}`).then(async(res)=>{
   image = await res.json();
   alert(image);
  })
}