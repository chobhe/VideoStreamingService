# READING FROM FLV FILE CODE
    
    # Ensure the file is in the right format
    # if not file.filename.endswith(".flv"):
    #     raise HTTPException(status_code=400, detail="Invalid file format")

    # Save the file to disk temporarily, you may also process it directly in memory for larger files
    # with open(f"/path/to/save/{file.filename}", "wb") as buffer:
    #     buffer.write(await file.read())

    # # Use OpenCV to process the video file
    # frames = []
    # cap = cv2.VideoCapture(f"/{file.filename}")
    # while cap.isOpened():
    #     ret, frame = cap.read()
    #     if not ret:
    #         break

    #     # frame is a numpy array containing an image
    #     frames.append(frame)

    # cap.release()

    # # Now you can process the frames with your computer vision code
    # # ...

    # return {"message": "Video processed successfully"}
