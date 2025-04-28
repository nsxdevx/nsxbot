package types

import "encoding/json"

type MeaasgeChain []Message

func (m MeaasgeChain) append(msg Message) MeaasgeChain {
	return append(m, msg)
}

func (m MeaasgeChain) Text(text string) MeaasgeChain {
	data, err := json.Marshal(Text{
		Text: text,
	})
	if err != nil {
		panic(err)
	}
	return m.append(Message{
		Type: "text",
		Data: data,
	})
}

func (m MeaasgeChain) Br() MeaasgeChain {
	return m.Text("\n")
}

func (m MeaasgeChain) Face(id string) MeaasgeChain {
	data, err := json.Marshal(Face{
		Id: id,
	})
	if err != nil {
		panic(err)
	}
	return m.append(Message{
		Type: "face",
		Data: data,
	})
}

func (m MeaasgeChain) At(qq string) MeaasgeChain {
	data, err := json.Marshal(At{
		QQ: qq,
	})
	if err != nil {
		panic(err)
	}
	return m.append(Message{
		Type: "at",
		Data: data,
	})
}

func (m MeaasgeChain) Reply(id int) MeaasgeChain {
	data, err := json.Marshal(Reply{
		Id: id,
	})
	if err != nil {
		panic(err)
	}
	return m.append(Message{
		Type: "reply",
		Data: data,
	})
}

// such as:
// network URL: https://www.example.com/image.png
// local file:///C:\\Users\Richard\Pictures\1.png see rfc 8089
// base64: base64://9j/4AAQSkZJRgABAQEAAAAAAAD/...
func (m MeaasgeChain) Image(file string) MeaasgeChain {
	data, err := json.Marshal(Image{
		BaseFile: BaseFile{
			File: file,
		},
	})
	if err != nil {
		panic(err)
	}
	return m.append(Message{
		Type: "image",
		Data: data,
	})
}

// such as:
// network URL: https://www.example.com/image.png
// local file:///C:\\Users\Richard\Pictures\1.png see rfc 8089
// base64: base64://9j/4AAQSkZJRgABAQEAAAAAAAD/...
func (m MeaasgeChain) File(file string) MeaasgeChain {
	data, err := json.Marshal(BaseFile{
		File: file,
	})
	if err != nil {
		panic(err)
	}
	return m.append(Message{
		Type: "file",
		Data: data,
	})
}
