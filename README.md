The Learning Management System (LMS) is designed to streamline educational processes and enhance the learning experience for both students and teachers. The system is controlled by an Admin who has the authority to manage users and system settings.The backend coding part is here. Golang is the programming language used to implement.

Admin Role
•	The admin is responsible for overall system control and management.
•	Admins can add teachers to the system by assigning them a username and password.

Teacher Role
•	Teachers can add, view, and update course materials for the courses they teach.
•	Only the teacher who owns a course has the authority to delete it. If a course deletion is required, a request is sent to the respective teacher via email.
•	Teachers receive requests from students wishing to join their courses. They can either accept or reject these requests, thereby managing student enrollment.
•	Teachers grade assignments submitted by students and assign marks accordingly.

Student Role
•	Students can register in the system. Upon registration, they receive a verification email, which they must confirm to proceed.
•	After email verification, students can browse available courses and send enrollment requests to the respective teachers.
•	Students can enroll in multiple courses, subject to the acceptance of their enrollment requests by the teachers.
•	Progress tracking is implemented, where students must complete the current task to unlock the next one.
•	Students must also complete assignments as part of their coursework.
•	The student's course progress is displayed at the top right corner of the page, allowing them to monitor their advancement.
