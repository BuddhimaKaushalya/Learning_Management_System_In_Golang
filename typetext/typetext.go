package typetext

// JSONB is an interface or type to handle custom struct that accepts more data in JSON format.
type JSONB map[string]interface{}

// Resource represents a resource with a title, description, and links.
type Resource struct {
	Title        string       `form:"title" json:"title"`
	Description  string       `form:"description" json:"description"`
	ResourceLink ResourceLink `json:"resource_link"`
}

// ResourceLink contains URLs related to the resource.
type ResourceLink struct {
	Url1 string `form:"url1" json:"url1"`
	Url2 string `form:"url2" json:"url2"`
	Url3 string `form:"url3" json:"url3"`
}

// WhatWill represents the learning objectives, skills gained, and target audience for a course.
type WhatWill struct {
	WhatWillYouLearn WhatWillYouLearn `json:"what_will_you_learn"`
	WhatSkillYouGain WhatSkillYouGain `json:"what_skill_you_gain"`
	WhoShouldLearn   WhoShouldLearn   `json:"who_should_learn"`
}

// WhatWillYouLearn describes the subjects covered in the course.
type WhatWillYouLearn struct {
	Subject1 string `form:"subject1" json:"subject1"`
	Subject2 string `form:"subject2" json:"subject2"`
	Subject3 string `form:"subject3" json:"subject3"`
}

// WhatSkillYouGain describes the skills gained from the course.
type WhatSkillYouGain struct {
	Skill1 string `form:"skill1" json:"skill1"`
	Skill2 string `form:"skill2" json:"skill2"`
	Skill3 string `form:"skill3" json:"skill3"`
}

// WhoShouldLearn describes the target audience for the course.
type WhoShouldLearn struct {
	Person1 string `form:"person1" json:"person1"`
	Person2 string `form:"person2" json:"person2"`
	Person3 string `form:"person3" json:"person3"`
}
